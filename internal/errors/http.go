package errors

import "net/http"

type HTTPError struct {
	statusCode int
	message    string
}

type HTTPCoder interface {
	HTTPCode() int
	error
}

func (e HTTPError) Error() string {
	return e.message
}

func (e HTTPError) HTTPCode() int {
	return e.statusCode
}

func (e HTTPError) Is(target error) bool {
	t, ok := target.(HTTPError)
	if !ok {
		return false
	}
	return e.statusCode == t.statusCode
}

func WriteHTTPError(w http.ResponseWriter, err error, defaultMsg string) {
	if coder, ok := err.(HTTPCoder); ok {
		http.Error(w, err.Error(), coder.HTTPCode())
		return
	}
	http.Error(w, defaultMsg, http.StatusInternalServerError)
}

func NotFound(msg string) HTTPError {
	return HTTPError{statusCode: http.StatusNotFound, message: msg}
}

func BadRequest(msg string) HTTPError {
	return HTTPError{statusCode: http.StatusBadRequest, message: msg}
}

func Unauthorized(msg string) HTTPError {
	return HTTPError{statusCode: http.StatusUnauthorized, message: msg}
}

func Forbidden(msg string) HTTPError {
	return HTTPError{statusCode: http.StatusForbidden, message: msg}
}

func Internal(msg string) HTTPError {
	return HTTPError{statusCode: http.StatusInternalServerError, message: msg}
}
