package middleware

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/xp/management-service/middleware/mocks"
)

func TestAuthorizationMiddleware(t *testing.T) {
	// Set up test handler that responds with request body and success status
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		_, _ = io.WriteString(w, string(body))
	})

	// Set up mock enforcer
	ok, nOk := true, false
	authzEnforcer := &mocks.Enforcer{}
	authzEnforcer.On(
		"UpsertPolicy",
		"allow-all-list-segmenters",
		[]string{},
		[]string{"**"},
		[]string{"segmenters"},
		[]string{"actions:read"},
	).Return(nil, nil)
	authzEnforcer.On(
		"UpsertPolicy",
		"validation-policy",
		[]string{},
		[]string{"**"},
		[]string{"validate"},
		[]string{"actions:create"},
	).Return(nil, nil)
	authzEnforcer.On("Enforce", "test-user@gojek.com", "projects", "actions:read").Return(&ok, nil)
	authzEnforcer.On("Enforce", "test-user@gojek.com", "projects:1", "actions:read").Return(&nOk, nil)
	authzEnforcer.On("Enforce", "test-user@gojek.com", "projects", "actions:update").Return(&nOk, nil)
	authzEnforcer.On("Enforce", "test-user-2@gojek.com", "projects", "actions:read").Return(&nOk, nil)

	// Create Authorizer
	authz, err := NewAuthorizer(authzEnforcer)
	assert.NoError(t, err)
	mw := authz.Middleware(testHandler)

	// Define tests
	tests := map[string]struct {
		method       string
		url          string
		body         string
		email        string
		expectedErr  bool
		expectedBody string
	}{
		"success": {
			method:       "GET",
			url:          "/projects",
			body:         `{"test": "value"}`,
			email:        "test-user@gojek.com",
			expectedBody: `{"test": "value"}`,
		},
		"failure | bad action": {
			method:       "PUT",
			url:          "/projects",
			body:         "{}",
			email:        "test-user@gojek.com",
			expectedErr:  true,
			expectedBody: `{"error":"test-user@gojek.com is not authorized to execute actions:update on projects"}`,
		},
		"failure | bad resource": {
			method:       "GET",
			url:          "/projects/1",
			body:         "{}",
			email:        "test-user@gojek.com",
			expectedErr:  true,
			expectedBody: `{"error":"test-user@gojek.com is not authorized to execute actions:read on projects:1"}`,
		},
		"failure | bad user": {
			method:       "GET",
			url:          "/projects",
			body:         "{}",
			email:        "test-user-2@gojek.com",
			expectedErr:  true,
			expectedBody: `{"error":"test-user-2@gojek.com is not authorized to execute actions:read on projects"}`,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(data.method, data.url, strings.NewReader(data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Email", data.email)

			// Create test response recorder and handle the request
			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			// Validate
			if data.expectedErr {
				assert.Equal(t, http.StatusUnauthorized, rr.Code)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
			assert.Equal(t, data.expectedBody, rr.Body.String())
		})
	}
}
