package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gojek/xp/common/api/schema"
)

const xpRequestIDHeaderKey = "XP-Request-ID"

// Response contains the status code and data to return to the caller
type Response struct {
	code int
	data interface{}
}

// WriteTo writes a Response to the provided http.ResponseWriter
func (r *Response) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(r.code)

	if r.data != nil {
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(r.data)
	}
}

func ErrorResponse(w http.ResponseWriter, statusCode int, err error, requestId *string) {
	w.Header().Set("Content-Type", "application/json")
	if requestId != nil {
		w.Header().Set(xpRequestIDHeaderKey, *requestId)
	}
	w.WriteHeader(statusCode)
	response := schema.Error{
		Code:    strconv.Itoa(statusCode),
		Message: err.Error(),
		Error:   err.Error(),
	}
	_ = json.NewEncoder(w).Encode(response)
}

func Ok(w http.ResponseWriter, jsonBody interface{}, requestId *string) {
	w.Header().Set("Content-Type", "application/json")
	if requestId != nil {
		w.Header().Set(xpRequestIDHeaderKey, *requestId)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jsonBody)
}
