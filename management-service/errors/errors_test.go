package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewfGetType(t *testing.T) {
	err := Newf(BadInput, "New %s error", "test")

	// Test that the error object has been created as expected
	assert.Equal(t, "New test error", err.Error())
	assert.Equal(t, BadInput, GetType(err))
}

func TestWrapfKnownNestedError(t *testing.T) {
	err := Newf(BadInput, "Inner error")
	wrappedErr := Wrapf(err, "Outer error %s", "message")

	assert.Equal(t, BadInput, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "Outer error message: Inner error")
}

func TestWrapfUnknownError(t *testing.T) {
	err := errors.New("Inner error")
	wrappedErr := Wrapf(err, "Outer error %s", "message")

	assert.Equal(t, Unknown, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "Outer error message: Inner error")
}

func TestGetErrorTypeUnknown(t *testing.T) {
	err := errors.New("Test error")
	assert.Equal(t, Unknown, GetType(err))
}

func TestGetHTTPErrorCode(t *testing.T) {
	testErrorSuite := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "Generic Error",
			err:          errors.New("Test error"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "BadInput",
			err:          Newf(BadInput, ""),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			err:          Newf(NotFound, ""),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, data := range testErrorSuite {
		t.Run(data.name, func(t *testing.T) {
			assert.Equal(t, data.expectedCode, GetHTTPErrorCode(data.err))
		})
	}
}

func TestNewHTTPError(t *testing.T) {
	message := "Test Error Message"
	err := fmt.Errorf(message)
	httpErr := NewHTTPError(err)
	// Unknown error code
	assert.Equal(t, message, httpErr.Message)
	assert.Equal(t, message, httpErr.Error())
	assert.Equal(t, 500, httpErr.Code)
	// Known error code
	err = Newf(BadInput, message)
	httpErr = NewHTTPError(err)
	assert.Equal(t, 400, httpErr.Code)
}
