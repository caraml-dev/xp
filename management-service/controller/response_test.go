package controller

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/turing-experiments/common/api/schema"
	"github.com/gojek/turing-experiments/management-service/errors"
)

func TestWriteErrorResponse(t *testing.T) {
	tests := map[string]struct {
		err              error
		expectedResponse string
		expectedStatus   int
	}{
		"generic error": {
			err:              fmt.Errorf("Test Internal Server Error"),
			expectedResponse: "Test Internal Server Error",
			expectedStatus:   500,
		},
		"bad input": {
			err:              errors.Newf(errors.BadInput, "Test Bad Request"),
			expectedResponse: "Test Bad Request",
			expectedStatus:   400,
		},
		"not found": {
			err:              errors.Newf(errors.NotFound, "Test Not Found"),
			expectedResponse: "Test Not Found",
			expectedStatus:   404,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteErrorResponse(w, data.err)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, _ := io.ReadAll(resp.Body)

			// Validate
			assert.Equal(t, data.expectedStatus, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.JSONEq(t,
				fmt.Sprintf(`{"code":"%d", "error":"%s", "message":"%s"}`, data.expectedStatus, data.expectedResponse, data.expectedResponse),
				string(body),
			)
		})
	}
}

func TestOk(t *testing.T) {
	paging1 := &schema.Paging{Page: 1, Pages: 2, Total: 3}
	paging2 := &schema.Paging{Page: 2, Pages: 2, Total: 4}

	tests := map[string]struct {
		jsonBody interface{}
		paging   []*schema.Paging
		expected string
	}{
		"no paging": {
			jsonBody: interface{}(map[string]string{"key": "val"}),
			paging:   []*schema.Paging{},
			expected: `{"data": {"key":"val"}}`,
		},
		"nil paging": {
			jsonBody: interface{}(map[string]string{"key": "val"}),
			paging:   []*schema.Paging{nil},
			expected: `{"data": {"key":"val"}}`,
		},
		"single paging": {
			jsonBody: interface{}(map[string]string{"key": "val"}),
			paging:   []*schema.Paging{paging1},
			expected: `{"data": {"key":"val"}, "paging": {"page": 1, "pages": 2, "total": 3}}`,
		},
		"multiple paging": { // Test that only the first paging object is used
			jsonBody: interface{}(map[string]string{"key": "val"}),
			paging:   []*schema.Paging{paging1, paging2},
			expected: `{"data": {"key":"val"}, "paging": {"page": 1, "pages": 2, "total": 3}}`,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Ok(w, data.jsonBody, data.paging...)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, _ := io.ReadAll(resp.Body)

			// Validate
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.JSONEq(t, data.expected, string(body))
		})
	}

}
