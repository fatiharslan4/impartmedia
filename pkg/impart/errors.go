package impart

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNotImplemented is returned when the requested method has not been implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrNotFound is returned when the requested resource was not found
var ErrNotFound = errors.New("requested resource not found")

// ErrBadRequest is returned when the request could not be completed due to an error in the request itself
var ErrBadRequest = errors.New("unable to complete the request")

// ErrUnknown is returned when the server encountered an error preventing the request from being completed
var ErrUnknown = errors.New("unknown server error")

// ErrUnauthorized is returned when the request could not be completed because the principal was not authorized
var ErrUnauthorized = errors.New("unauthorized")

// ErrMethodNotSupported is returned when the requested HttpMethod is not supported
var ErrMethodNotSupported = errors.New("http method not supported")

// ErrExists is returned when attempting to create a resource that already exists
var ErrExists = errors.New("resource already exists")

// ErrNoOp is returned when attempting to update something to a state that is already in the proper state.
var ErrNoOp = errors.New("resource already matches exactly as request")

var ErrNoAPIKey = errors.New("no api key provided")

type Error interface {
	error
	HttpStatus() int
	ToJson() string
	Err() error
	Msg() string
}

var _ error = &impartError{}

type impartError struct {
	err error
	msg string
}

func (e impartError) Error() string {
	return e.err.Error()
}

func (e impartError) Err() error {
	return e.err
}

func (e impartError) Msg() string {
	return e.msg
}

func NewError(err error, msg string) Error {
	return impartError{err: err, msg: msg}
}

var UnknownError = NewError(ErrUnknown, "Internal Server Error")

// ErrorCheck takes an input error and returns a formatted api gateway response
func (e impartError) HttpStatus() int {
	var statusCode int
	switch e.err {
	case ErrBadRequest:
		statusCode = http.StatusBadRequest
	case ErrNotFound:
		statusCode = http.StatusNotFound
	case ErrNotImplemented:
		statusCode = http.StatusNotImplemented
	case ErrUnknown:
		statusCode = http.StatusInternalServerError
	case ErrMethodNotSupported:
		statusCode = http.StatusMethodNotAllowed
	case ErrUnauthorized:
		statusCode = http.StatusUnauthorized
	case ErrExists:
		statusCode = http.StatusConflict
	default:
		statusCode = http.StatusInternalServerError
	}
	return statusCode
}
func (e impartError) ToJson() string {
	type S struct {
		Err string `json:"error"`
		Msg string `json:"msg"`
	}

	b, _ := json.Marshal(S{e.err.Error(), e.msg})
	return string(b)
}

func (e impartError) MarshalJSON() ([]byte, error) {
	return []byte(e.ToJson()), nil
}
