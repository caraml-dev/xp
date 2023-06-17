package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/xp/management-service/api"
)

func TestNewOpenAPIValidator(t *testing.T) {
	swagger1, err := api.GetSwagger()
	require.NoError(t, err)
	swagger2, _ := api.GetSwagger()
	swagger2.Servers = nil
	tests := map[string]struct {
		ignoreAuth          bool
		ignoreServers       bool
		expectedSwagger     *openapi3.T
		expectedNilAuthFunc bool
	}{
		"default": {
			expectedSwagger:     swagger1,
			expectedNilAuthFunc: true,
		},
		"ignore auth and servers": {
			ignoreAuth:      true,
			ignoreServers:   true,
			expectedSwagger: swagger2,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := NewOpenAPIValidator(&OpenAPIValidationOptions{
				IgnoreAuthentication: data.ignoreAuth,
				IgnoreServers:        data.ignoreServers,
			})
			// Validate
			require.NoError(t, err)
			assert.Equal(t, data.expectedSwagger, val.swagger)
			assert.Equal(t, data.expectedNilAuthFunc, val.options.AuthenticationFunc == nil)
		})
	}
}

func TestOpenAPIMiddleware(t *testing.T) {
	// Set up test handler that responds with request body and success status
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		_, _ = io.WriteString(w, string(body))
	})

	// Set up openapi middleware
	val, err := NewOpenAPIValidator(&OpenAPIValidationOptions{
		IgnoreAuthentication: true,
		IgnoreServers:        true,
	})
	require.NoError(t, err)
	mw := val.Middleware()

	// Define tests
	tests := map[string]struct {
		method       string
		url          string
		body         string
		expectedErr  bool
		expectedBody string
	}{
		"get projects": {
			method: "GET",
			url:    "/projects",
		},
		"get project settings": {
			method:      "GET",
			url:         "/projects/abc/settings",
			expectedErr: true,
			expectedBody: strings.Join([]string{"parameter \"project_id\" in path has an error: value abc: an invalid integer: ",
				"strconv.ParseFloat: parsing \"abc\": invalid syntax\n"}, ""),
		},
		"get experiment variables": {
			method: "GET",
			url:    "/projects/1/experiment-variables",
		},
		"create project settings | no rand key": {
			method: "POST",
			url:    "/projects/1/settings",
			body: `{
				"segmenters": {
					"names": ["seg1"],
					"variables": {
					  "seg1": ["seg1"]
					}
				}
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/randomization_key\": property \"randomization_key\" is missing\n",
		},
		"create project settings | no segmenter": {
			method: "POST",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1"
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters\": property \"segmenters\" is missing\n",
		},
		"create project settings | no segmenter mappings": {
			method: "POST",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1",
				"segmenters": {
					"names": ["seg1"]
				}
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters/variables\": property \"variables\" is missing\n",
		},
		"create project settings | no segmenter names": {
			method: "POST",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1",
				"segmenters": {
					"variables": {
					  "seg1": ["seg1"]
					}
				}
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters/names\": property \"names\" is missing\n",
		},
		"update project settings| no rand key": {
			method: "PUT",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": 1,
				"segmenters": {
					"names": ["seg1"],
					"variables": {
					  "seg1": ["seg1"]
					}
				}
			}`,
			expectedErr: true,
			expectedBody: strings.Join([]string{"request body has an error: doesn't match the schema: Error at \"/randomization_key\": ",
				"Field must be set to string or not be present\n"}, ""),
		},
		"update project settings | no segmenter": {
			method: "PUT",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1"
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters\": property \"segmenters\" is missing\n",
		},
		"update project settings | no variables": {
			method: "PUT",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1",
				"segmenters": {
					"names": ["seg1"]
				}
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters/variables\": property \"variables\" is missing\n",
		},
		"update project settings | no segmenters names": {
			method: "PUT",
			url:    "/projects/1/settings",
			body: `{
				"randomization_key": "1",
				"segmenters": {
					"variables": {
					  "seg1": ["seg1"]
					}
				}
			}`,
			expectedErr:  true,
			expectedBody: "request body has an error: doesn't match the schema: Error at \"/segmenters/names\": property \"names\" is missing\n",
		},
		"get experiments": {
			method: "GET",
			url:    "/projects/2/experiments",
		},
		"get experiment": {
			method: "GET",
			url:    "/projects/2/experiments/3",
		},
		"create experiment": {
			method: "POST",
			url:    "/projects/2/experiments",
			body: `{
				"name": "abc",
				"start_time": "2021-03-01T00:00:00Z",
				"end_time": "2021-03-01T00:00:00Z",
				"type": "unknown_type",
				"status": "active"
			}`,
			expectedErr: true,
			expectedBody: strings.Join([]string{"request body has an error: doesn't match the schema: Error at \"/type\": ",
				"value is not one of the allowed values\n"}, ""),
		},
		"unknown url": {
			method:       "GET",
			url:          "/unknown",
			body:         `{}`,
			expectedErr:  true,
			expectedBody: "no matching operation was found\n",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(data.method, data.url, strings.NewReader(data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create test response recorder and handle the request
			rr := httptest.NewRecorder()
			mw(testHandler).ServeHTTP(rr, req)

			// Validate
			if data.expectedErr {
				assert.NotEqual(t, http.StatusOK, rr.Code)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
			assert.Equal(t, data.expectedBody, rr.Body.String())
		})
	}
}
