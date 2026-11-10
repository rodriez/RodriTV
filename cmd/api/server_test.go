package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeValidationError struct {
	msg    string
	fields map[string]string
}

func (e fakeValidationError) Error() string { return e.msg }
func (e fakeValidationError) ValidationFields() map[string]string {
	return e.fields
}

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func TestWriteError_Mapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantFields map[string]string
	}{
		{
			name:       "validation -> 400",
			err:        fakeValidationError{msg: "name is required", fields: map[string]string{"name": "is required"}},
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation",
			wantFields: map[string]string{"name": "is required"},
		},
		{
			name:       "wrapped validation -> 400",
			err:        fmtWrap(fakeValidationError{msg: "name is required"}),
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation",
		},
		{
			name:       "not found -> 404",
			err:        errNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name:       "conflict -> 409",
			err:        errConflict,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict",
		},
		{
			name:       "generic -> 500",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeError(rec, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var body errorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantCode, body.Error.Code)
			assert.NotEmpty(t, body.Error.Message)
			assert.Equal(t, tt.wantFields, body.Error.Fields)
		})
	}
}

func fmtWrap(err error) error {
	return fmt.Errorf("create channel: %w", err)
}

func TestWriteError_GenericDoesNotLeakMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, errors.New("db password is s3cret"))

	var body errorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal server error", body.Error.Message)
	assert.NotContains(t, rec.Body.String(), "s3cret")
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var gotID string
	handler := requestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = r.Context().Value(ctxKeyRequestID{}).(string)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.NotEmpty(t, gotID)
	assert.Equal(t, gotID, rec.Header().Get("X-Request-ID"))
}

func TestRequestID_PropagatesIncoming(t *testing.T) {
	var gotID string
	handler := requestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = r.Context().Value(ctxKeyRequestID{}).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "req-123", gotID)
	assert.Equal(t, "req-123", rec.Header().Get("X-Request-ID"))
}

func TestLogging_WritesStructuredLine(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	handler := requestID(logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))

	req := httptest.NewRequest(http.MethodPost, "/v1/channels", nil)
	req.Header.Set("X-Request-ID", "req-abc")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "req-abc", line["request_id"])
	assert.Equal(t, "POST", line["method"])
	assert.Equal(t, "/v1/channels", line["path"])
	assert.InDelta(t, http.StatusCreated, line["status"], 0)
	assert.Contains(t, line, "duration_ms")
}

func TestRecoverPanic_Returns500AndLogsStack(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	handler := recoverPanic(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var body errorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal", body.Error.Code)

	assert.Contains(t, buf.String(), "kaboom")
	assert.Contains(t, buf.String(), "goroutine")
}

func TestNewServer_ChainServesHealthz(t *testing.T) {
	var buf bytes.Buffer
	srv := NewServer(Config{Port: 9291}, newTestLogger(&buf), fakePinger{})

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "GET", line["method"])
	assert.Equal(t, "/healthz", line["path"])
	assert.InDelta(t, http.StatusOK, line["status"], 0)
}
