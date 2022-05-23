package controller

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOk(t *testing.T) {
	requestId := "1"
	w := httptest.NewRecorder()
	Ok(w, interface{}(map[string]string{"key": "val"}), &requestId)
	resp := w.Result()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "1", resp.Header.Get(xpRequestIDHeaderKey))
	assert.JSONEq(t, `{"key":"val"}`, string(body))
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	requestId := "2"
	ErrorResponse(w, http.StatusBadRequest, errors.New("bad request"), &requestId)
	resp := w.Result()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "2", resp.Header.Get(xpRequestIDHeaderKey))
	assert.JSONEq(t, `{"code":"400", "error":"bad request", "message":"bad request"}`, string(body))
}

func TestWriteTo(t *testing.T) {
	w := httptest.NewRecorder()
	resp := Response{
		code: 200,
		data: map[string]string{"key": "val"},
	}
	resp.WriteTo(w)
	copiedResp := w.Result()
	if copiedResp != nil && copiedResp.Body != nil {
		defer copiedResp.Body.Close()
	}
	body, _ := io.ReadAll(copiedResp.Body)

	assert.Equal(t, 200, copiedResp.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", copiedResp.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"key":"val"}`, string(body))
}
