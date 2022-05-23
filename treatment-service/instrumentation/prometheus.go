package instrumentation

import (
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Namespace is the Prometheus Namespace in all metrics published by the xp app
	Namespace string = "mlp"
	// Subsystem is the Prometheus Subsystem in all metrics published by the xp app
	Subsystem string = "xp_treatment_service"
	// FetchTreatmentRequestDurationMs is the key to measure http requests for fetching treatment
	FetchTreatmentRequestDurationMs metrics.MetricName = "fetch_treatment_request_duration_ms"
	// ExperimentLookupDurationMs is the key to measure experiment lookup duration
	ExperimentLookupDurationMs metrics.MetricName = "experiment_lookup_duration_ms"
	// FetchTreatmentRequestCount is the key to measure no. of fetch treatment requests
	FetchTreatmentRequestCount metrics.MetricName = "fetch_treatment_request_count"
	// NoMatchingExperimentRequestCount is the key to measure no. of fetch treatment requests with no matching experiments
	NoMatchingExperimentRequestCount metrics.MetricName = "no_matching_experiment_request_count"
)

// requestLatencyBuckets defines the buckets used in the custom Histogram metrics
var requestLatencyBuckets = []float64{
	5, 10, 15, 20, 30, 40, 50, 100, 200, 500, 1000,
}

var GaugeMap = map[metrics.MetricName]metrics.PrometheusGaugeVec{}

func GetCounterMap(labels []string) map[metrics.MetricName]metrics.PrometheusCounterVec {
	allLabels := append(
		labels, "project_name", "experiment_name", "treatment_name", "response_code",
	)
	noMatchingExperimentlabels := append(
		labels, "project_name", "response_code",
	)

	counterMap := map[metrics.MetricName]metrics.PrometheusCounterVec{
		FetchTreatmentRequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Help:      "Counter for no. of Fetch Treatment requests with matching experiments",
			Name:      string(FetchTreatmentRequestCount),
		},
			allLabels,
		),
		NoMatchingExperimentRequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Help:      "Counter for no. of Fetch Treatment requests with no matching experiments",
			Name:      string(NoMatchingExperimentRequestCount),
		},
			noMatchingExperimentlabels,
		),
	}

	return counterMap
}

func GetHistogramMap() map[metrics.MetricName]metrics.PrometheusHistogramVec {

	histogramMap := map[metrics.MetricName]metrics.PrometheusHistogramVec{
		FetchTreatmentRequestDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(FetchTreatmentRequestDurationMs),
			Help:      "Histogram for the runtime (in milliseconds) of Fetch Treatment requests",
			Buckets:   requestLatencyBuckets,
		},
			[]string{"project_name", "experiment_name", "treatment_name", "response_code"},
		),
		ExperimentLookupDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(ExperimentLookupDurationMs),
			Help:      "Histogram for the runtime (in milliseconds) of experiment lookup",
			Buckets:   requestLatencyBuckets,
		},
			[]string{"project_name"},
		),
	}

	return histogramMap
}
