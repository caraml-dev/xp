package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gojek/xp/common/api/schema"
)

func NotFound(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	response := schema.Error{
		Code:    "404",
		Message: err.Error(),
	}
	_ = json.NewEncoder(w).Encode(response)
}

func BadRequest(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	response := schema.Error{
		Code:    "400",
		Message: err.Error(),
	}
	_ = json.NewEncoder(w).Encode(response)
}

func Success(w http.ResponseWriter, jsonBody interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jsonBody)
}
