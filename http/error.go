package http

import "net/http"

type appError struct {
	Error       error
	Message     string
	Code        int
	ContentType string
}

func internalServerError(err error) *appError {
	return &appError{
		Error:   err,
		Message: "Internal server error",
		Code:    http.StatusInternalServerError,
	}
}

func notFound(err error) *appError {
	return &appError{Error: err, Code: http.StatusNotFound}
}

func badRequest(err error) *appError {
	return &appError{Error: err, Code: http.StatusBadRequest}
}

func (e *appError) AsJSON() *appError {
	e.ContentType = jsonMediaType
	return e
}

func (e *appError) WithMessage(message string) *appError {
	e.Message = message
	return e
}

func (e *appError) IsJSON() bool {
	return e.ContentType == jsonMediaType
}
