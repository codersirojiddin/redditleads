package http

import "net/http"

// AppError is a structured error that maps directly to an HTTP response.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func NewAppError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func BadRequest(msg string) *AppError { return NewAppError(http.StatusBadRequest, "bad_request", msg) }
func Unauthorized(msg string) *AppError {
	return NewAppError(http.StatusUnauthorized, "unauthorized", msg)
}
func Forbidden(msg string) *AppError { return NewAppError(http.StatusForbidden, "forbidden", msg) }
func NotFound(msg string) *AppError  { return NewAppError(http.StatusNotFound, "not_found", msg) }
func Conflict(msg string) *AppError  { return NewAppError(http.StatusConflict, "conflict", msg) }
func Internal(msg string) *AppError {
	return NewAppError(http.StatusInternalServerError, "internal_error", msg)
}

// WriteError renders an error (AppError or generic) as a JSON response.
func WriteError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*AppError); ok {
		JSON(w, appErr.Status, appErr)
		return
	}
	JSON(w, http.StatusInternalServerError, &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "something went wrong",
	})
}
