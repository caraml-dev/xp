package middleware

import (
	"net/http"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"

	"github.com/gojek/xp/common/utils"
)

func NewRelicMiddleware() func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			_, instrumentedHandler := newrelic.WrapHandle(utils.GetRoutePattern(r), h)
			instrumentedHandler.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
