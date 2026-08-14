package instrumentation

import (
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
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
	// FetchTreatmentRequestDurationMsHelpString is the help string of the FetchTreatmentRequestDurationMs metric
	FetchTreatmentRequestDurationMsHelpString string = "Histogram for the runtime (in milliseconds) of Fetch Treatment requests"
	// ExperimentLookupDurationMsHelpString is the help string of the ExperimentLookupDurationMs metric
	ExperimentLookupDurationMsHelpString string = "Histogram for the runtime (in milliseconds) of experiment lookup"
	// FetchTreatmentRequestCountHelpString is the help string of the FetchTreatmentRequestCount metric
	FetchTreatmentRequestCountHelpString string = "Counter for no. of Fetch Treatment requests with matching experiments"
	// NoMatchingExperimentRequestCountHelpString is the help string of the NoMatchingExperimentRequestCount metric
	NoMatchingExperimentRequestCountHelpString string = "Counter for no. of Fetch Treatment requests with no matching experiments"
	// LocalStorageCallCount is the key to measure calls to LocalStorage's public methods
	LocalStorageCallCount metrics.MetricName = "local_storage_calls_total"
	// LocalStorageCallDurationMs is the key to measure how long a call to a LocalStorage public method took
	LocalStorageCallDurationMs metrics.MetricName = "local_storage_call_duration_ms"
	// LocalStorageCallCountHelpString is the help string of the LocalStorageCallCount metric
	LocalStorageCallCountHelpString string = "Counter for calls to LocalStorage's public methods, incremented on entry"
	// LocalStorageCallDurationMsHelpString is the help string of the LocalStorageCallDurationMs metric
	LocalStorageCallDurationMsHelpString string = "Histogram for how long (in milliseconds) a call to a LocalStorage public method took to return"
)

// RequestLatencyBuckets defines the buckets used in the custom Histogram metrics
var RequestLatencyBuckets = []float64{
	5, 10, 15, 20, 30, 40, 50, 100, 200, 500, 1000,
}

// AdditionalFetchTreatmentRequestCountLabels defines additional labels needed for the FetchTreatmentRequestCount
// counter map
var AdditionalFetchTreatmentRequestCountLabels = []string{"project_name", "experiment_name", "treatment_name",
	"response_code"}

// AdditionalNoMatchingExperimentRequestCountLabels defines additional labels needed for the NoMatchingExperimentRequestCount
// counter map
var AdditionalNoMatchingExperimentRequestCountLabels = []string{"project_name", "response_code"}

// FetchTreatmentRequestDurationMsLabels defines additional labels needed for the FetchTreatmentRequestDurationMs
// histogram map
var FetchTreatmentRequestDurationMsLabels = []string{"project_name", "experiment_name", "treatment_name", "response_code"}

// ExperimentLookupDurationMsLabels defines additional labels needed for the ExperimentLookupDurationMs histogram map
var ExperimentLookupDurationMsLabels = []string{"project_name"}

// LocalStorageMethodLabels defines the labels needed for the LocalStorageCallCount counter map and the
// LocalStorageCallDurationMs histogram map. Deliberately not combined with the caller-supplied custom
// MetricLabels (cfg.MetricLabels) used elsewhere in this file -- LocalStorage calls aren't tied to a project
// or experiment, just a method name.
var LocalStorageMethodLabels = []string{"method"}

var GaugeMap = map[metrics.MetricName]metrics.PrometheusGaugeVec{}

func GetCounterMap(labels []string) map[metrics.MetricName]metrics.PrometheusCounterVec {
	allLabels := append(
		labels, AdditionalFetchTreatmentRequestCountLabels...,
	)
	noMatchingExperimentlabels := append(
		labels, AdditionalNoMatchingExperimentRequestCountLabels...,
	)

	counterMap := map[metrics.MetricName]metrics.PrometheusCounterVec{
		FetchTreatmentRequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Help:      FetchTreatmentRequestCountHelpString,
			Name:      string(FetchTreatmentRequestCount),
		},
			allLabels,
		),
		NoMatchingExperimentRequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Help:      NoMatchingExperimentRequestCountHelpString,
			Name:      string(NoMatchingExperimentRequestCount),
		},
			noMatchingExperimentlabels,
		),
		LocalStorageCallCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Help:      LocalStorageCallCountHelpString,
			Name:      string(LocalStorageCallCount),
		},
			LocalStorageMethodLabels,
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
			Help:      FetchTreatmentRequestDurationMsHelpString,
			Buckets:   RequestLatencyBuckets,
		},
			FetchTreatmentRequestDurationMsLabels,
		),
		ExperimentLookupDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(ExperimentLookupDurationMs),
			Help:      ExperimentLookupDurationMsHelpString,
			Buckets:   RequestLatencyBuckets,
		},
			ExperimentLookupDurationMsLabels,
		),
		LocalStorageCallDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(LocalStorageCallDurationMs),
			Help:      LocalStorageCallDurationMsHelpString,
			Buckets:   RequestLatencyBuckets,
		},
			LocalStorageMethodLabels,
		),
	}

	return histogramMap
}
