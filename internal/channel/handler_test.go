package channel

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// testWriteError replica el mapeo uniforme de cmd/api/server.go para
// verificar el contrato de error completo desde los tests del handler.
func testWriteError(w http.ResponseWriter, err error) {
	var verr ErrValidation

	status := http.StatusInternalServerError
	body := errorResponse{Error: errorBody{Code: "internal", Message: "internal server error"}}

	switch {
	case errors.As(err, &verr):
		status = http.StatusBadRequest
		body.Error = errorBody{Code: "validation", Message: verr.Error(), Fields: verr.ValidationFields()}
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		body.Error = errorBody{Code: "not_found", Message: err.Error()}
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
		body.Error = errorBody{Code: "conflict", Message: err.Error()}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newTestHandler(repo *mockRepository) *Handler {
	return NewHandler(NewService(repo, time.Now), testWriteError)
}

func doRequest(t *testing.T, handler http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var resp errorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	return resp
}

func TestHandlerList(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		ch.Sources[0] = &Source{ID: uuid.New(), Slot: 1, URL: "https://example.com/live.m3u8", Enabled: true}
		repo.On("List", mock.Anything).Return([]Channel{ch}, nil)

		rec := doRequest(t, newTestHandler(repo).List, http.MethodGet, "/v1/channels", "")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var body struct {
			Channels []map[string]json.RawMessage `json:"channels"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		require.Len(t, body.Channels, 1)
		assert.JSONEq(t, `"El Trece"`, string(body.Channels[0]["name"]))

		var sources []json.RawMessage
		require.NoError(t, json.Unmarshal(body.Channels[0]["sources"], &sources))
		require.Len(t, sources, 5)
		assert.JSONEq(t, `null`, string(sources[1]))
		assert.JSONEq(t, `null`, string(sources[4]))
		repo.AssertExpectations(t)
	})

	t.Run("empty list serializes as []", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("List", mock.Anything).Return([]Channel{}, nil)

		rec := doRequest(t, newTestHandler(repo).List, http.MethodGet, "/v1/channels", "")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"channels":[]}`, rec.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("List", mock.Anything).Return([]Channel(nil), errors.New("db down"))

		rec := doRequest(t, newTestHandler(repo).List, http.MethodGet, "/v1/channels", "")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "internal", decodeError(t, rec).Error.Code)
	})
}

func TestHandlerCreate(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		repo.On("ExistsNumber", mock.Anything, ch.Number, uuid.Nil).Return(false, nil)
		repo.On("Create", mock.Anything, mock.Anything).Return(ch, nil)

		body := `{"number":1,"name":"El Trece","sources":[null,null,null,null,null]}`
		rec := doRequest(t, newTestHandler(repo).Create, http.MethodPost, "/v1/channels", body)

		require.Equal(t, http.StatusCreated, rec.Code)
		var created Channel
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&created))
		assert.Equal(t, ch.ID, created.ID)
		repo.AssertExpectations(t)
	})

	t.Run("duplicate number returns 409 conflict", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("ExistsNumber", mock.Anything, 1, uuid.Nil).Return(true, nil)

		body := `{"number":1,"name":"El Trece"}`
		rec := doRequest(t, newTestHandler(repo).Create, http.MethodPost, "/v1/channels", body)

		require.Equal(t, http.StatusConflict, rec.Code)
		resp := decodeError(t, rec)
		assert.Equal(t, "conflict", resp.Error.Code)
	})

	t.Run("invalid body returns 400 validation", func(t *testing.T) {
		rec := doRequest(t, newTestHandler(new(mockRepository)).Create, http.MethodPost, "/v1/channels", `{invalid`)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		resp := decodeError(t, rec)
		assert.Equal(t, "validation", resp.Error.Code)
		assert.Contains(t, resp.Error.Fields, "body")
	})

	t.Run("domain validation returns 400 validation", func(t *testing.T) {
		body := `{"number":0,"name":"El Trece"}`
		rec := doRequest(t, newTestHandler(new(mockRepository)).Create, http.MethodPost, "/v1/channels", body)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		resp := decodeError(t, rec)
		assert.Equal(t, "validation", resp.Error.Code)
		assert.Contains(t, resp.Error.Fields, "number")
	})
}

func TestHandlerGet(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		repo.On("GetByID", mock.Anything, ch.ID).Return(ch, nil)

		handler := http.HandlerFunc(newTestHandler(repo).Get)
		req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+ch.ID.String(), nil)
		req.SetPathValue("id", ch.ID.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var got Channel
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
		assert.Equal(t, ch.ID, got.ID)
	})

	t.Run("not found returns 404 not_found", func(t *testing.T) {
		repo := new(mockRepository)
		id := uuid.New()
		repo.On("GetByID", mock.Anything, id).Return(Channel{}, ErrNotFound)

		handler := http.HandlerFunc(newTestHandler(repo).Get)
		req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+id.String(), nil)
		req.SetPathValue("id", id.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "not_found", decodeError(t, rec).Error.Code)
	})

	t.Run("malformed id returns 400 validation", func(t *testing.T) {
		handler := http.HandlerFunc(newTestHandler(new(mockRepository)).Get)
		req := httptest.NewRequest(http.MethodGet, "/v1/channels/not-a-uuid", nil)
		req.SetPathValue("id", "not-a-uuid")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		resp := decodeError(t, rec)
		assert.Equal(t, "validation", resp.Error.Code)
		assert.Contains(t, resp.Error.Fields, "id")
	})
}

func TestHandlerUpdate(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		repo.On("ExistsNumber", mock.Anything, ch.Number, ch.ID).Return(false, nil)
		repo.On("Update", mock.Anything, mock.MatchedBy(func(updated Channel) bool {
			return updated.ID == ch.ID
		})).Return(ch, nil)

		handler := http.HandlerFunc(newTestHandler(repo).Update)
		body := `{"number":1,"name":"El Trece","segment_seconds":5}`
		req := httptest.NewRequest(http.MethodPut, "/v1/channels/"+ch.ID.String(), strings.NewReader(body))
		req.SetPathValue("id", ch.ID.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		repo.AssertExpectations(t)
	})

	t.Run("malformed id returns 400", func(t *testing.T) {
		handler := http.HandlerFunc(newTestHandler(new(mockRepository)).Update)
		req := httptest.NewRequest(http.MethodPut, "/v1/channels/bad", strings.NewReader(`{}`))
		req.SetPathValue("id", "bad")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("duplicate number returns 409", func(t *testing.T) {
		repo := new(mockRepository)
		id := uuid.New()
		repo.On("ExistsNumber", mock.Anything, 1, id).Return(true, nil)

		handler := http.HandlerFunc(newTestHandler(repo).Update)
		body := `{"number":1,"name":"El Trece","segment_seconds":5}`
		req := httptest.NewRequest(http.MethodPut, "/v1/channels/"+id.String(), strings.NewReader(body))
		req.SetPathValue("id", id.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
		assert.Equal(t, "conflict", decodeError(t, rec).Error.Code)
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		handler := http.HandlerFunc(newTestHandler(new(mockRepository)).Update)
		req := httptest.NewRequest(http.MethodPut, "/v1/channels/"+uuid.NewString(), strings.NewReader(`{bad`))
		req.SetPathValue("id", uuid.NewString())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandlerDelete(t *testing.T) {
	t.Run("no content", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		repo.On("GetByID", mock.Anything, ch.ID).Return(ch, nil)
		repo.On("Delete", mock.Anything, ch.ID).Return(nil)

		handler := http.HandlerFunc(newTestHandler(repo).Delete)
		req := httptest.NewRequest(http.MethodDelete, "/v1/channels/"+ch.ID.String(), nil)
		req.SetPathValue("id", ch.ID.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
		repo.AssertExpectations(t)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		repo := new(mockRepository)
		id := uuid.New()
		repo.On("GetByID", mock.Anything, id).Return(Channel{}, ErrNotFound)

		handler := http.HandlerFunc(newTestHandler(repo).Delete)
		req := httptest.NewRequest(http.MethodDelete, "/v1/channels/"+id.String(), nil)
		req.SetPathValue("id", id.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "not_found", decodeError(t, rec).Error.Code)
	})

	t.Run("malformed id returns 400", func(t *testing.T) {
		handler := http.HandlerFunc(newTestHandler(new(mockRepository)).Delete)
		req := httptest.NewRequest(http.MethodDelete, "/v1/channels/bad", nil)
		req.SetPathValue("id", "bad")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
