package tvchannels

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ApiError struct {
	Status int      `json:"-"`
	URL    string   `json:"type,omitempty"`
	Title  string   `json:"title,omitempty"`
	Detail string   `json:"detail,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

func (apiErr *ApiError) write(w http.ResponseWriter) {
	w.WriteHeader(apiErr.Status)
	w.Header().Add("Content-Type", "applicaiton/json")

	_ = json.NewEncoder(w).Encode(apiErr)
}

func FromError(err error) ApiError {
	if errors.As(err, new(ValidationError)) {
		return ApiErrBadRequest("One or more field are invalid", err.Error())
	} else if errors.As(err, new(MissingResourceError)) {
		return ApiErrNotFound("Resource not exist", err.Error())
	} else {
		return ApiErrInternal("Internal error")
	}
}

func NewApiError(status int, detail string, errors ...string) ApiError {
	return ApiError{
		Status: status,
		Title:  http.StatusText(status),
		Detail: detail,
		Errors: errors,
	}
}

func ApiErrBadRequest(detail string, errors ...string) ApiError {
	return NewApiError(http.StatusBadRequest, detail, errors...)
}

func ApiErrInternal(detail string, errors ...string) ApiError {
	return NewApiError(http.StatusInternalServerError, detail, errors...)
}

func ApiErrNotFound(detail string, errors ...string) ApiError {
	return NewApiError(http.StatusNotFound, detail, errors...)
}
