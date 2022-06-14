package middleware

import (
	"context"
	"fmt"
	"net/http"

	mw "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/gojek/xp/management-service/api"
)

type OpenAPIValidationOptions struct {
	// If true, ignore "security" in OpenAPI specs
	IgnoreAuthentication bool
	// If true, ignore "server" declarations in OpenAPI specs when validating requests paths.
	// Only consider the paths relative to the server url versus checking the full paths
	// (which include the server URL) in the requests.
	IgnoreServers bool
}

type OpenAPIValidator struct {
	swagger *openapi3.T
	options openapi3filter.Options
}

func NewOpenAPIValidator(options *OpenAPIValidationOptions) (*OpenAPIValidator, error) {
	// Get Swagger specs
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("Error loading swagger spec\n: %s", err)
	}

	// Handle options
	openAPIFilterOpts := openapi3filter.Options{}
	if options != nil {
		if options.IgnoreAuthentication {
			openAPIFilterOpts.AuthenticationFunc = func(
				context.Context,
				*openapi3filter.AuthenticationInput,
			) error {
				return nil
			}
		}
		if options.IgnoreServers {
			swagger.Servers = nil
		}
	}
	return &OpenAPIValidator{swagger, openAPIFilterOpts}, nil
}

func (v *OpenAPIValidator) Middleware() func(http.Handler) http.Handler {
	return mw.OapiRequestValidatorWithOptions(v.swagger, &mw.Options{Options: v.options})
}
