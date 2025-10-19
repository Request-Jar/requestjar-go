package errors

import (
	"errors"
	"net/http"
)

var (
	New    = errors.New
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
)

var (
	ErrNotFound     = HTTPError{statusCode: http.StatusNotFound, message: "not found"}
	ErrBadRequest   = HTTPError{statusCode: http.StatusBadRequest, message: "bad request"}
	ErrUnauthorized = HTTPError{statusCode: http.StatusUnauthorized, message: "unauthorized"}
	ErrForbidden    = HTTPError{statusCode: http.StatusForbidden, message: "unauthorized"}
	ErrInternal     = HTTPError{statusCode: http.StatusInternalServerError, message: "unauthorized"}
)
