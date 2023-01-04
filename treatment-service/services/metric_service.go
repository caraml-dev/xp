package services

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/caraml-dev/xp/common/api/schema"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/instrumentation"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

type MetricService interface {
	LogFetchTreatmentMetrics(
		begin time.Time,
		projectId models.ProjectId,
		treatment schema.SelectedTreatment,
		requestFilter map[string][]*_segmenters.SegmenterValue,
		statusCode int,
	)
	LogLatencyHistogram(begin time.Time, labels map[string]string, loggingMetric metrics.MetricName)
	LogRequestCount(labels map[string]string, loggingMetric metrics.MetricName)

	// GetProjectNameLabel retrieves only project name as labels
	GetProjectNameLabel(projectId models.ProjectId) map[string]string
	GetMetricLabels() []string
	// GetLabels retrieves labels with flag to filter for segmenters
	GetLabels(projectId models.ProjectId, treatment schema.SelectedTreatment, statusCode int,
		requestFilter map[string][]*_segmenters.SegmenterValue, withSegmenters bool) map[string]string
	SetMetricsCollector(collector metrics.Collector)
}

type metricService struct {
	Kind         config.MetricSinkKind
	LocalStorage *models.LocalStorage
	MetricLabels []string
}

func NewMetricService(cfg config.Monitoring, localStorage *models.LocalStorage) (MetricService, error) {
	switch cfg.Kind {
	case config.NoopMetricSink, config.RPCMetricSink:
	case config.PrometheusMetricSink:
		// Init metrics collector
		histogramMap := instrumentation.GetHistogramMap()
		counterMap := instrumentation.GetCounterMap(cfg.MetricLabels)
		err := metrics.InitPrometheusMetricsCollector(instrumentation.GaugeMap, histogramMap, counterMap)
		if err != nil {
			return nil, errors.New("failed to initialize Prometheus-based MetricService")
		}
	}

	svc := &metricService{
		Kind:         cfg.Kind,
		LocalStorage: localStorage,
		MetricLabels: cfg.MetricLabels,
	}

	return svc, nil
}

func (ms *metricService) SetMetricsCollector(collector metrics.Collector) {
	metrics.SetGlobMetricsCollector(collector)
	return
}

func (ms *metricService) LogLatencyHistogram(begin time.Time, labels map[string]string, loggingMetric metrics.MetricName) {
	var err error
	switch ms.Kind {
	case config.NoopMetricSink:
	case config.PrometheusMetricSink, config.RPCMetricSink:
		switch loggingMetric {
		case instrumentation.FetchTreatmentRequestDurationMs:
			err = metrics.Glob().MeasureDurationMsSince(
				instrumentation.FetchTreatmentRequestDurationMs, begin, labels,
			)
		case instrumentation.ExperimentLookupDurationMs:
			err = metrics.Glob().MeasureDurationMsSince(
				instrumentation.ExperimentLookupDurationMs, begin, labels,
			)
		}
		if err != nil {
			log.Printf("error while logging %s metrics (latency): %s", loggingMetric, err)
		}
	}
}

func (ms *metricService) LogRequestCount(labels map[string]string, loggingMetric metrics.MetricName) {
	var err error
	switch ms.Kind {
	case config.NoopMetricSink:
	case config.PrometheusMetricSink, config.RPCMetricSink:
		switch loggingMetric {
		case instrumentation.FetchTreatmentRequestCount:
			err = metrics.Glob().Inc(
				instrumentation.FetchTreatmentRequestCount, labels,
			)
		case instrumentation.NoMatchingExperimentRequestCount:
			err = metrics.Glob().Inc(
				instrumentation.NoMatchingExperimentRequestCount, labels,
			)
		}
		if err != nil {
			log.Printf("error while logging metrics (request_count): %s", err)
		}
	}
}

func (ms *metricService) GetProjectNameLabel(projectId models.ProjectId) map[string]string {
	settings := ms.LocalStorage.FindProjectSettingsWithId(projectId)
	return map[string]string{
		"project_name": settings.Username,
	}
}

func (ms *metricService) GetMetricLabels() []string {
	return ms.MetricLabels
}

func (ms *metricService) GetLabels(
	projectId models.ProjectId,
	treatment schema.SelectedTreatment,
	statusCode int,
	requestFilter map[string][]*_segmenters.SegmenterValue,
	withSegmenters bool,
) map[string]string {
	settings := ms.LocalStorage.FindProjectSettingsWithId(projectId)
	labels := map[string]string{
		"project_name":    settings.Username,
		"experiment_name": treatment.ExperimentName,
		"treatment_name":  treatment.Treatment.Name,
		"response_code":   strconv.Itoa(statusCode),
	}

	if withSegmenters {
		// Set default value for required labels
		for _, label := range ms.MetricLabels {
			labels[label] = ""
			if filterValues, ok := requestFilter[label]; ok {
				// Do the convert and set labels[label]
				strLabels := []string{}
				for _, v := range filterValues {
					switch v.Value.(type) {
					case *_segmenters.SegmenterValue_Integer:
						strLabels = append(strLabels, strconv.Itoa(int(v.GetInteger())))
					case *_segmenters.SegmenterValue_String_:
						strLabels = append(strLabels, v.GetString_())
					}
				}
				labels[label] = strings.Join(strLabels, ",")
			}
		}
	}
	return labels
}

func (ms *metricService) LogFetchTreatmentMetrics(
	begin time.Time,
	projectId models.ProjectId,
	treatment schema.SelectedTreatment,
	requestFilter map[string][]*_segmenters.SegmenterValue,
	statusCode int,
) {
	labels := ms.GetLabels(
		projectId,
		treatment,
		statusCode,
		requestFilter,
		false,
	)
	ms.LogLatencyHistogram(begin, labels, instrumentation.FetchTreatmentRequestDurationMs)

	labels = ms.GetLabels(
		projectId,
		treatment,
		statusCode,
		requestFilter,
		true,
	)
	if treatment.ExperimentName != "" || treatment.Treatment.Name != "" {
		ms.LogRequestCount(labels, instrumentation.FetchTreatmentRequestCount)
	} else {
		delete(labels, "experiment_name")
		delete(labels, "treatment_name")
		ms.LogRequestCount(labels, instrumentation.NoMatchingExperimentRequestCount)
	}
}
