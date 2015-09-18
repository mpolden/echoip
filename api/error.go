package api

import "net/http"

type appError struct {
	Error       error
	Response    string
	Code        int
	ContentType string
}

func internalServerError(err error) *appError {
	return &appError{Error: err, Response: "Internal server error", Code: http.StatusInternalServerError}
}

func notFound(err error) *appError {
	return &appError{Error: err, Code: http.StatusNotFound}
}

func (e *appError) WithContentType(contentType string) *appError {
	e.ContentType = contentType
	return e
}

func (e *appError) WithCode(code int) *appError {
	e.Code = code
	return e
}

func (e *appError) WithResponse(response string) *appError {
	e.Response = response
	return e
}

func (e *appError) WithError(err error) *appError {
	e.Error = err
	return e
}

func (e *appError) IsJSON() bool {
	return e.ContentType == APPLICATION_JSON
}
