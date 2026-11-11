package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLogoRequest(t *testing.T, filename string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/logos/"+filename, nil)
	req.SetPathValue("filename", filename)

	return req
}

func writeLogo(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("fake-image"), 0o644))
}

type logoErrorBody struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields"`
	} `json:"error"`
}

func TestFSLogoStore_ServeExistingLogo(t *testing.T) {
	dir := t.TempDir()
	writeLogo(t, dir, "3f6b2c1e.png")

	rec := httptest.NewRecorder()
	NewFSLogoStore(dir).ServeHTTP(rec, newLogoRequest(t, "3f6b2c1e.png"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, "fake-image", rec.Body.String())
}

func TestFSLogoStore_ServeContentTypeByExtension(t *testing.T) {
	dir := t.TempDir()
	store := NewFSLogoStore(dir)

	for name, contentType := range map[string]string{
		"a.jpg":  "image/jpeg",
		"a.jpeg": "image/jpeg",
		"a.webp": "image/webp",
		"a.svg":  "image/svg+xml",
	} {
		writeLogo(t, dir, name)

		rec := httptest.NewRecorder()
		store.ServeHTTP(rec, newLogoRequest(t, name))

		assert.Equal(t, http.StatusOK, rec.Code, name)
		assert.Equal(t, contentType, rec.Header().Get("Content-Type"), name)
	}
}

func TestFSLogoStore_ServePathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	store := NewFSLogoStore(dir)

	for _, name := range []string{"..%2F..%2Fetc%2Fpasswd", "../etc/passwd", "a/b.png", "a\\b.png", ".."} {
		rec := httptest.NewRecorder()
		store.ServeHTTP(rec, newLogoRequest(t, name))

		assert.Equal(t, http.StatusBadRequest, rec.Code, name)

		var body logoErrorBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "validation", body.Error.Code, name)
		assert.Contains(t, body.Error.Fields, "filename", name)
	}
}

func TestFSLogoStore_ServeDisallowedExtension(t *testing.T) {
	dir := t.TempDir()
	writeLogo(t, dir, "evil.exe")

	rec := httptest.NewRecorder()
	NewFSLogoStore(dir).ServeHTTP(rec, newLogoRequest(t, "evil.exe"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var body logoErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "validation", body.Error.Code)
}

func TestFSLogoStore_ServeNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	NewFSLogoStore(t.TempDir()).ServeHTTP(rec, newLogoRequest(t, "missing.png"))

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var body logoErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_found", body.Error.Code)
}

func TestFSLogoStore_Delete(t *testing.T) {
	dir := t.TempDir()
	writeLogo(t, dir, "logo.png")
	store := NewFSLogoStore(dir)

	require.NoError(t, store.Delete(context.Background(), "logo.png"))
	_, err := os.Stat(filepath.Join(dir, "logo.png"))
	require.ErrorIs(t, err, os.ErrNotExist)

	// Borrado idempotente: archivo inexistente no es error.
	require.NoError(t, store.Delete(context.Background(), "logo.png"))

	// Nombre inválido devuelve el error de validación de dominio.
	err = store.Delete(context.Background(), "../x.png")
	var verr ErrValidation
	require.ErrorAs(t, err, &verr)
	assert.Equal(t, "filename", verr.Field)
}

func TestFSLogoStore_SatisfiesLogoStoreInterface(t *testing.T) {
	var store LogoStore = NewFSLogoStore(t.TempDir())
	require.NoError(t, store.Delete(context.Background(), "missing.png"))
}
