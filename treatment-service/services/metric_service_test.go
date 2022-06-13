package services

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gojek/turing-experiments/common/api/schema"
	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/treatment-service/config"
	"github.com/gojek/turing-experiments/treatment-service/instrumentation"
	"github.com/gojek/turing-experiments/treatment-service/internal/testutils"
	"github.com/gojek/turing-experiments/treatment-service/models"
)

type MetricServiceTestSuite struct {
	suite.Suite
	MetricService
	storage models.LocalStorage
	cfg     config.Config
}

func newProjectSettings(
	enableS2idClustering bool,
	passkey string,
	projectId int64,
	randomizationKey string,
	segmenterNames []string,
	segmentersMap map[string][]string,
	username string,
) schema.ProjectSettings {
	createdAt := time.Now()
	variables := schema.ProjectSegmenters_Variables{AdditionalProperties: segmentersMap}
	segmenters := schema.ProjectSegmenters{
		Names:     segmenterNames,
		Variables: variables,
	}

	return schema.ProjectSettings{
		CreatedAt:            createdAt,
		EnableS2idClustering: enableS2idClustering,
		Passkey:              passkey,
		ProjectId:            projectId,
		RandomizationKey:     randomizationKey,
		Segmenters:           segmenters,
		UpdatedAt:            createdAt,
		Username:             username,
	}
}

func (s *MetricServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up MetricServiceTestSuite")

	addProjectSettings := func(projectSettings schema.ProjectSettings) {
		e := models.OpenAPIProjectSettingsSpecToProtobuf(projectSettings)
		s.storage.ProjectSettings = append(s.storage.ProjectSettings, e)
	}
	segmenterNames := []string{"days_of_week", "hours_of_day"}
	segmenters := map[string][]string{
		"days_of_week": {"tz"},
		"hours_of_day": {"tz"},
	}
	addProjectSettings(newProjectSettings(
		false, "passkey", 0, "randomkey", segmenterNames, segmenters, "user1"))

	var err error
	s.cfg = config.Config{
		MonitoringConfig: config.Monitoring{
			Kind:         config.PrometheusMetricSink,
			MetricLabels: []string{"test_label1", "test_label2"},
		},
	}
	s.MetricService, err = NewMetricService(s.cfg.MonitoringConfig, &s.storage)
	if err != nil {
		s.Suite.T().Log("failed to initialize MetricService")
	}
}

func (s *MetricServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up MetricServiceTestSuite")
}

func TestMetricService(t *testing.T) {
	suite.Run(t, new(MetricServiceTestSuite))
}

func (s *MetricServiceTestSuite) TestGetLabels() {
	projectId := models.NewProjectId(0)
	treatment := schema.SelectedTreatment{
		ExperimentId:   0,
		ExperimentName: "test_exp",
		Treatment: schema.SelectedTreatmentData{
			Name: "test_treatment",
		},
	}
	statusCode := 200
	metricLabels := []string{}
	requestFilter := map[string][]*_segmenters.SegmenterValue{}

	labels := s.MetricService.GetLabels(projectId, treatment, statusCode, metricLabels, requestFilter, false)
	expectedLabels := map[string]string{
		"project_name":    "user1",
		"experiment_name": "test_exp",
		"treatment_name":  "test_treatment",
		"response_code":   strconv.Itoa(statusCode),
	}

	s.Suite.Require().Equal(expectedLabels, labels)
}

func (s *MetricServiceTestSuite) TestGetLabelsWithSegmenters() {
	projectId := models.NewProjectId(0)
	treatment := schema.SelectedTreatment{
		ExperimentId:   0,
		ExperimentName: "test_exp",
		Treatment: schema.SelectedTreatmentData{
			Name: "test_treatment",
		},
	}
	statusCode := 200
	metricLabels := s.cfg.MonitoringConfig.MetricLabels
	requestFilter := map[string][]*_segmenters.SegmenterValue{
		"test_extra_label": {{Value: &_segmenters.SegmenterValue_String_{String_: "test"}}},
	}

	labels := s.MetricService.GetLabels(projectId, treatment, statusCode, metricLabels, requestFilter, true)
	expectedLabels := map[string]string{
		"project_name":    "user1",
		"experiment_name": "test_exp",
		"treatment_name":  "test_treatment",
		"response_code":   strconv.Itoa(statusCode),
		"test_label1":     "",
		"test_label2":     "",
	}

	s.Suite.Require().Equal(expectedLabels, labels)
}

func (s *MetricServiceTestSuite) TestLogLatencyHistogram() {
	label := map[string]string{
		"project_name": "test",
	}
	invalid := map[string]string{
		"dummy": "test",
	}

	stdout := testutils.CaptureStderrLogs(func() {
		s.MetricService.LogLatencyHistogram(time.Now(), label, instrumentation.ExperimentLookupDurationMs)
	})
	s.Suite.Require().Equal("", stdout)

	stdout = testutils.CaptureStderrLogs(func() {
		s.MetricService.LogLatencyHistogram(time.Now(), invalid, instrumentation.ExperimentLookupDurationMs)
	})
	expectedErrorStdOut := "error while logging experiment_lookup_duration_ms metrics (latency)"
	s.Suite.Require().Contains(stdout, expectedErrorStdOut)
}

func (s *MetricServiceTestSuite) TestLogRequestCount() {
	label := map[string]string{
		"project_name":    "user1",
		"experiment_name": "test_exp",
		"treatment_name":  "test_treatment",
		"response_code":   strconv.Itoa(200),
		"test_label1":     "",
		"test_label2":     "",
	}
	stdout := testutils.CaptureStderrLogs(func() {
		s.MetricService.LogRequestCount(label, instrumentation.FetchTreatmentRequestCount)
	})
	s.Suite.Require().Equal("", stdout)

	stdout = testutils.CaptureStderrLogs(func() {
		s.MetricService.LogRequestCount(map[string]string{}, instrumentation.FetchTreatmentRequestCount)
	})
	expectedErrorStdOut := "error while logging metrics (request_count)"
	s.Suite.Require().Contains(stdout, expectedErrorStdOut)
}
