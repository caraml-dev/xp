package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/mocks"
)

type ExperimentHistoryControllerTestSuite struct {
	suite.Suite
	ctrl                              *ExperimentHistoryController
	expectedExperimentHistoryResponse string
	expectedErrorResponseFormat       string
}

func (s *ExperimentHistoryControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ExperimentHistoryControllerTestSuite")

	// Create mock MLP service and set up with test responses
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On(
		"GetProject", int64(1),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(1)))
	mlpSvc.On("GetProject", int64(2)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(3)).Return(nil, nil)

	// Create mock project settings service and set up with test responses
	settingsSvc := &mocks.ProjectSettingsService{}
	settingsSvc.
		On("GetDBRecord", models.ID(2)).
		Return(nil, errors.Newf(errors.Unknown, "test get project settings error"))
	settingsSvc.
		On("GetDBRecord", models.ID(3)).
		Return(nil, nil)

	// Create mock experiment service and set up with test responses
	expSvc := &mocks.ExperimentService{}
	expSvc.
		On("GetDBRecord", models.ID(3), models.ID(1)).
		Return(nil, errors.Newf(errors.NotFound, "experiment not found"))
	expSvc.
		On("GetDBRecord", models.ID(3), models.ID(10)).
		Return(nil, nil)

	// Set up mock experiment history service
	var testExperimentTraffic, testExperimentInterval int32 = 100, 10
	var testDescription = "test-description"
	var testName = "control"
	var testConfig = map[string]interface{}{
		"config-1": "value",
		"config-2": 2,
	}
	testExpHistory := &models.ExperimentHistory{
		Model: models.Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:           models.ID(100),
		ExperimentID: models.ID(40),
		Version:      int64(8),
		Name:         "exp-hist-1",
		Description:  &testDescription,
		Type:         models.ExperimentTypeSwitchback,
		Tier:         models.ExperimentTierDefault,
		Interval:     &testExperimentInterval,
		Treatments: []models.ExperimentTreatment{
			{
				Configuration: testConfig,
				Name:          testName,
				Traffic:       &testExperimentTraffic,
			},
		},
		Segment: models.ExperimentSegment{
			"days_of_week": []string{"1"},
		},
		Status:    models.ExperimentStatusInactive,
		EndTime:   time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC),
		StartTime: time.Date(2022, 2, 2, 1, 1, 1, 0, time.UTC),
		UpdatedBy: "test-updated-by",
	}
	expHistSvc := &mocks.ExperimentHistoryService{}
	expHistSvc.
		On("GetExperimentHistory", int64(10), int64(2)).
		Return(nil, errors.Newf(errors.NotFound, "experiment history not found"))
	expHistSvc.
		On("GetExperimentHistory", int64(10), int64(1)).
		Return(testExpHistory, nil)
	expHistSvc.
		On("ListExperimentHistory", int64(10), services.ListExperimentHistoryParams{
			PaginationOptions: pagination.PaginationOptions{},
		}).
		Return([]*models.ExperimentHistory{testExpHistory}, &pagination.Paging{Page: 1, Total: 1, Pages: 1}, nil)

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("GetSegmenterTypes", int64(3)).
		Return(
			map[string]schema.SegmenterType{
				"days_of_week": schema.SegmenterTypeInteger,
			},
			nil,
		)

	// Set up expected responses
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedExperimentHistoryResponse = `{
		"created_at": "2021-01-01T02:03:04Z",
		"description": "test-description",
		"end_time": "2022-01-01T01:01:01Z",
		"updated_at": "2021-01-01T02:03:04Z",
		"experiment_id": 40,
		"id": 100,
		"interval": 10,
		"name": "exp-hist-1",
		"segment": {
			"days_of_week": [1]
		},
		"start_time": "2022-02-02T01:01:01Z",
		"status":  "inactive",
		"tier": "default",
		"treatments": [{
			"configuration": {
				"config-1": "value",
				"config-2": 2
			},
			"name": "control",
			"traffic": 100
		}],
		"type": "Switchback",
		"updated_by": "test-updated-by",
		"version": 8
	}`

	// Create test controller
	s.ctrl = &ExperimentHistoryController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				ExperimentService:        expSvc,
				ExperimentHistoryService: expHistSvc,
				MLPService:               mlpSvc,
				ProjectSettingsService:   settingsSvc,
				SegmenterService:         segmenterSvc,
			},
		},
	}
}

func TestExperimentHistoryController(t *testing.T) {
	suite.Run(t, new(ExperimentHistoryControllerTestSuite))
}

func (s *ExperimentHistoryControllerTestSuite) TestGetExperimentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name         string
		projectID    int64
		experimentID int64
		version      int64
		expected     string
	}{
		{
			name:         "mlp project not found",
			projectID:    1,
			experimentID: 1,
			version:      1,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:         "project settings not found",
			projectID:    2,
			experimentID: 1,
			version:      1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:         "experiment not found",
			projectID:    3,
			experimentID: 1,
			version:      1,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Experiment with id 1 cannot be retrieved: experiment not found\""),
		},
		{
			name:         "experiment history not found",
			projectID:    3,
			experimentID: 10,
			version:      2,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"experiment history not found\""),
		},
		{
			name:         "success",
			projectID:    3,
			experimentID: 10,
			version:      1,
			expected:     fmt.Sprintf(`{"data": %s}`, s.expectedExperimentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetExperimentHistory(w, nil, data.projectID, data.experimentID, data.version)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (s *ExperimentHistoryControllerTestSuite) TestListExperimentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name         string
		projectID    int64
		experimentID int64
		params       api.ListExperimentHistoryParams
		expected     string
	}{
		{
			name:         "mlp project not found",
			projectID:    1,
			experimentID: 1,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:         "project settings not found",
			projectID:    2,
			experimentID: 1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:         "experiment not found",
			projectID:    3,
			experimentID: 1,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Experiment with id 1 cannot be retrieved: experiment not found\""),
		},
		{
			name:         "success",
			projectID:    3,
			experimentID: 10,
			expected:     fmt.Sprintf(`{"data": [%s], "paging": {"page": 1, "pages": 1, "total": 1}}`, s.expectedExperimentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(nil))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListExperimentHistory(w, req, data.projectID, data.experimentID, data.params)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}
