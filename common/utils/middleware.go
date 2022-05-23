package utils

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Relevant thread: https://github.com/go-chi/chi/issues/270#issuecomment-479184559
func GetRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if pattern := rctx.RoutePattern(); pattern != "" {
		// Pattern is already available
		return pattern
	}

	routePath := r.URL.Path
	if r.URL.RawPath != "" {
		routePath = r.URL.RawPath
	}

	tctx := chi.NewRouteContext()
	if !rctx.Routes.Match(tctx, r.Method, routePath) {
		// No matching pattern, so just return the request path.
		// Depending on your use case, it might make sense to
		// return an empty string or error here instead
		return routePath
	}

	// tctx has the updated pattern, since Match mutates it
	return tctx.RoutePattern()
}
