package models

import (
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
	"github.com/caraml-dev/xp/treatment-service/instrumentation"
)

// LocalStorageMetricsRecorder is implemented by whatever wants to observe calls to
// LocalStorage's public methods. It's deliberately just services.MetricService's existing
// LogRequestCount/LogLatencyHistogram methods, rather than bespoke ones -- a MetricService
// satisfies this as-is. models has no import on services (which itself imports models for
// *LocalStorage, e.g. in NewMetricService), so appcontext wires the two together after both
// are constructed, via SetMetricsRecorder.
type LocalStorageMetricsRecorder interface {
	LogRequestCount(labels map[string]string, loggingMetric metrics.MetricName)
	LogLatencyHistogram(begin time.Time, labels map[string]string, loggingMetric metrics.MetricName)
}

// instrumentCall reports a call to method via s.metricsRecorder (a no-op if none is set) and
// returns a func that reports the call's duration; callers defer the returned func
// immediately, e.g.:
//
//	func (s *LocalStorage) Foo(...) ... {
//		if s.metricsRecorder != nil {
//			defer s.instrumentCall("Foo")()
//		}
//		...
//	}
func (s *LocalStorage) instrumentCall(method string) func() {
	if s.metricsRecorder == nil {
		return func() {}
	}

	labels := map[string]string{"method": method}
	s.metricsRecorder.LogRequestCount(labels, instrumentation.LocalStorageCallCount)
	begin := time.Now()
	return func() {
		s.metricsRecorder.LogLatencyHistogram(begin, labels, instrumentation.LocalStorageCallDurationMs)
	}
}
