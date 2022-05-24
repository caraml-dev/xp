// Package errors provides definitions for some common errors that may be encountered.
// These errors may be eventually mapped to HTTP error codes, when being returned as a response.
package errors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// ErrorType captures some common error types
type ErrorType uint

const (
	// Unknown error type is used for all generic errors
	Unknown = ErrorType(iota)
	// BadInput is used when any function encounters bad/incomplete input
	BadInput
	// NotFound is used when a resource cannot be located
	NotFound
)

type errorData struct {
	Type ErrorType
	Info error
}

// Error satisfies error interface
func (error errorData) Error() string {
	return error.Info.Error()
}

// Newf creates a new errorData of the specified type with formatted message
func Newf(et ErrorType, msg string, args ...interface{}) error {
	err := fmt.Errorf(msg, args...)
	return errorData{Type: et, Info: err}
}

// Wrapf creates a new wrapped errorData with formatted message
func Wrapf(err error, msg string, args ...interface{}) error {
	newErr := errors.Wrapf(err, msg, args...)
	// Try casting the inner error to errorData
	if errData, ok := err.(errorData); ok {
		return errorData{
			Type: errData.Type,
			Info: newErr,
		}
	}
	return errorData{Type: Unknown, Info: newErr}
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if errData, ok := err.(errorData); ok {
		return errData.Type
	}
	return Unknown
}

// GetHTTPErrorCode maps the ErrorType to http status codes and returns it
func GetHTTPErrorCode(err error) int {
	var code int

	et := GetType(err)

	switch et {
	case BadInput:
		code = http.StatusBadRequest
	case NotFound:
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}
	return code
}

// HTTPError associates an error message with a HTTP status code.
type HTTPError struct {
	Code    int
	Message string
}

// Error satisfies the error interface
func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(err error) *HTTPError {
	return &HTTPError{
		Code:    GetHTTPErrorCode(err),
		Message: err.Error(),
	}
}
