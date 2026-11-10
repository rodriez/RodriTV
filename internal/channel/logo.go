package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// logoNamePattern restringe los nombres de logo a caracteres seguros,
// rechazando separadores de ruta (path traversal).
var logoNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// logoContentTypes mapea las extensiones permitidas a su Content-Type.
var logoContentTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
}

// FSLogoStore almacena y sirve los logos de canales desde el filesystem,
// confinados al directorio dir. Satisface la interfaz LogoStore del Service
// y http.Handler para el endpoint GET /v1/logos/{filename}.
type FSLogoStore struct {
	dir string
}

// NewFSLogoStore crea un almacén de logos sobre el directorio dir.
func NewFSLogoStore(dir string) FSLogoStore {
	return FSLogoStore{dir: dir}
}

// ServeHTTP sirve `${dir}/{filename}` tras validar el nombre: responde
// 400 (validation) si el nombre es inválido o la extensión no es de imagen,
// y 404 si el archivo no existe. El cuerpo de error sigue el formato
// uniforme {"error":{code,message,fields}} de la API.
func (s FSLogoStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("filename")
	if err := validateLogoName(name); err != nil {
		writeLogoError(w, http.StatusBadRequest, "validation", err)
		return
	}

	path := filepath.Join(s.dir, name)
	if _, err := os.Stat(path); err != nil {
		writeLogoError(w, http.StatusNotFound, "not_found", fmt.Errorf("logo %q not found", name))
		return
	}

	w.Header().Set("Content-Type", logoContentTypes[strings.ToLower(filepath.Ext(name))])
	http.ServeFile(w, r, path)
}

// Delete elimina el logo name del directorio. Devuelve ErrValidation si el
// nombre es inválido y nil si el archivo no existe (borrado idempotente).
func (s FSLogoStore) Delete(_ context.Context, name string) error {
	if err := validateLogoName(name); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(s.dir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete logo %q: %w", name, err)
	}

	return nil
}

// validateLogoName rechaza path traversal: solo [a-zA-Z0-9._-] y extensión
// de imagen permitida (png, jpg, jpeg, webp, svg).
func validateLogoName(name string) error {
	if !logoNamePattern.MatchString(name) {
		return ErrValidation{Field: "filename", Msg: "must match [a-zA-Z0-9._-]"}
	}
	if _, ok := logoContentTypes[strings.ToLower(filepath.Ext(name))]; !ok {
		return ErrValidation{Field: "filename", Msg: "extension must be one of png, jpg, jpeg, webp, svg"}
	}

	return nil
}

// writeLogoError escribe el cuerpo de error uniforme de la API
// ({"error":{code,message,fields}}) desde internal/channel, donde el
// helper writeError de cmd/api no es alcanzable.
func writeLogoError(w http.ResponseWriter, status int, code string, err error) {
	body := map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": err.Error(),
		},
	}

	var verr ErrValidation
	if errors.As(err, &verr) {
		body["error"].(map[string]any)["fields"] = verr.ValidationFields()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
