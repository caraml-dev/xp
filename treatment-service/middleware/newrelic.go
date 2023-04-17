package middleware

import (
	"net/http"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"

	"github.com/caraml-dev/xp/common/utils"
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
