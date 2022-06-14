package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/errors"
)

func Ok(w http.ResponseWriter, jsonBody interface{}, paging ...*schema.Paging) {
	w.Header().Set("Content-Type", "application/json")

	if jsonBody == nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	resp := struct {
		Data   interface{} `json:"data"`
		Paging interface{} `json:"paging,omitempty"`
	}{
		Data: jsonBody,
	}
	if len(paging) > 0 && paging[0] != nil {
		resp.Paging = paging[0]
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	httpErr := errors.NewHTTPError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Code)
	response := schema.Error{
		Code:    strconv.Itoa(httpErr.Code),
		Message: err.Error(),
		Error:   err.Error(),
	}
	_ = json.NewEncoder(w).Encode(response)
}
