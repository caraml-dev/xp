package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type ExperimentControllerTestSuite struct {
	suite.Suite
	ctrl                        *ExperimentController
	expectedExperimentResponses []string
	expectedErrorResponseFormat string
}

func (s *ExperimentControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ExperimentControllerTestSuite")

	// Create mock project settings service and set up with test responses
	settingsSvc := &mocks.ProjectSettingsService{}
	settingsSvc.
		On("GetDBRecord", models.ID(1)).
		Return(nil, errors.Newf(errors.Unknown, "test find project settings error"))
	settingsSvc.
		On("GetDBRecord", models.ID(2)).
		Return(&models.Settings{ProjectID: models.ID(2)}, nil)
	settingsSvc.
		On("GetProjectSettings", int64(1)).
		Return(nil, errors.Newf(errors.NotFound, "test get project settings error"))
	settingsSvc.
		On("GetProjectSettings", int64(2)).
		Return(&models.Settings{
			Config: &models.ExperimentationConfig{},
		}, nil)
	settingsSvc.
		On("GetProjectSettings", int64(3)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(5)).
		Return(&models.Settings{
			Config: &models.ExperimentationConfig{Segmenters: models.ProjectSegmenters{
				Names: []string{"days_of_week"},
			}},
		}, nil)

	// Create mock experiment service and set up with test responses
	expSvc := &mocks.ExperimentService{}
	testExperiment := &models.Experiment{ProjectID: 2}
	testExperiment1 := &models.Experiment{ProjectID: 2, Tier: models.ExperimentTierOverride}
	daysOfWeek := []string{"1", "2", "3", "4", "5", "6", "7"}
	testExperiment2 := &models.Experiment{ProjectID: 5, Segment: models.ExperimentSegment{
		"days_of_week": daysOfWeek,
	}}
	s.expectedExperimentResponses = []string{
		`{
			"project_id": 2,
			"created_at": "0001-01-01T00:00:00Z",
			"end_time": "0001-01-01T00:00:00Z",
			"id": 0,
			"name": "",
			"description": null,
			"interval": null,
			"segment": {},
			"treatments": null,
			"status": "",
			"tier": "",
			"type": "",
			"start_time": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z",
			"updated_by": ""
		}`,
		`{
			"project_id": 2,
			"created_at": "0001-01-01T00:00:00Z",
			"end_time": "0001-01-01T00:00:00Z",
			"id": 0,
			"name": "",
			"description": null,
			"interval": null,
			"segment": {},
			"treatments": null,
			"status": "",
			"tier": "override",
			"type": "",
			"start_time": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z",
			"updated_by": ""
		}`,
		`{
			"project_id": 5,
			"created_at": "0001-01-01T00:00:00Z",
			"end_time": "0001-01-01T00:00:00Z",
			"id": 0,
			"name": "",
			"description": null,
			"interval": null,
			"segment": {"days_of_week": [1,2,3,4,5,6,7]},
			"treatments": null,
			"status": "",
			"tier": "",
			"type": "",
			"start_time": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z",
			"updated_by": ""
		}`,
	}
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	expSvc.
		On("GetExperiment", int64(2), int64(20)).
		Return(nil, errors.Newf(errors.NotFound, "experiment not found"))
	expSvc.
		On("GetExperiment", int64(2), int64(2)).
		Return(testExperiment, nil)
	expSvc.
		On("GetExperiment", int64(5), int64(1)).
		Return(testExperiment2, nil)
	var emptyStatus *models.ExperimentStatus
	var emptyType *models.ExperimentType
	updatedBy := "test-user"
	expSvc.
		On("ListExperiments", int64(3), services.ListExperimentsParams{
			Status: emptyStatus, Type: emptyType, Segment: models.ExperimentSegment{},
		}).Return(nil, nil, fmt.Errorf("unexpected error"))
	expSvc.
		On("ListExperiments", int64(2), services.ListExperimentsParams{
			Status: emptyStatus, Type: emptyType, Segment: models.ExperimentSegment{"days_of_week": []string{"1"}},
		}).Return([]*models.Experiment{testExperiment}, nil, nil)
	expSvc.
		On("ListExperiments", int64(2),
			services.ListExperimentsParams{
				Status:  emptyStatus,
				Type:    emptyType,
				Segment: models.ExperimentSegment{}}).
		Return([]*models.Experiment{testExperiment}, nil, nil)
	expSvc.
		On("CreateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateExperimentRequestBody{
				Name:      "test-exp",
				UpdatedBy: &updatedBy,
				Tier:      models.ExperimentTierDefault,
				Segment:   models.ExperimentSegmentRaw(nil),
			}).
		Return(nil, fmt.Errorf("experiment creation failed"))
	expSvc.
		On("CreateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateExperimentRequestBody{
				Name:      "test-exp-2",
				UpdatedBy: &updatedBy,
				Tier:      models.ExperimentTierDefault,
				Segment:   models.ExperimentSegmentRaw(nil),
			}).
		Return(testExperiment, nil)
	expSvc.
		On("CreateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateExperimentRequestBody{
				Name:      "test-exp-2",
				UpdatedBy: &updatedBy,
				Tier:      models.ExperimentTierOverride,
				Segment:   models.ExperimentSegmentRaw(nil),
			}).
		Return(testExperiment1, nil)
	testDescription := "test-description-2"
	testDaysOfWeek := []interface{}{float64(1)}
	expSvc.
		On("UpdateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateExperimentRequestBody{
				Segment:   models.ExperimentSegmentRaw{"days_of_week": testDaysOfWeek},
				UpdatedBy: &updatedBy,
				Tier:      models.ExperimentTierDefault,
			}).
		Return(nil, fmt.Errorf("experiment update failed"))
	expSvc.
		On("UpdateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateExperimentRequestBody{
				Description: &testDescription,
				UpdatedBy:   &updatedBy,
				Tier:        models.ExperimentTierDefault,
				Segment:     models.ExperimentSegmentRaw(nil),
			}).
		Return(testExperiment, nil)
	expSvc.
		On("UpdateExperiment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateExperimentRequestBody{
				Description: &testDescription,
				UpdatedBy:   &updatedBy,
				Tier:        models.ExperimentTierOverride,
				Segment:     models.ExperimentSegmentRaw(nil),
			}).
		Return(testExperiment1, nil)
	expSvc.
		On("EnableExperiment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1)).
		Return(nil)
	expSvc.
		On("EnableExperiment",
			models.Settings{ProjectID: models.ID(2)},
			int64(3)).
		Return(errors.Newf(errors.BadInput, "experiment id 3 is already active"))
	expSvc.
		On("DisableExperiment",
			int64(2),
			int64(1)).
		Return(nil)
	expSvc.
		On("DisableExperiment",
			int64(2),
			int64(3)).
		Return(errors.Newf(errors.BadInput, "experiment id 3 is already inactive"))

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("ListSegmenterNames").
		Return([]string{"days_of_week"})
	segmenterSvc.
		On("GetSegmenterConfigurations",
			[]string{"days_of_week"}).
		Return(
			[]*_segmenters.SegmenterConfiguration{
				{Name: "days_of_week", Type: _segmenters.SegmenterValueType_INTEGER},
			}, nil,
		)
	segmenterSvc.
		On("GetSegmenterTypes").
		Return(
			map[string]schema.SegmenterType{
				"hours_of_day": schema.SegmenterTypeInteger,
				"days_of_week": schema.SegmenterTypeInteger,
			},
		)

	// Create mock MLP service and set up with test responses
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(2)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(3)).Return(nil, nil)
	mlpSvc.On(
		"GetProject", int64(4),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(4)))
	mlpSvc.On("GetProject", int64(5)).Return(nil, nil)

	// Create test controller
	s.ctrl = &ExperimentController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				ExperimentService:      expSvc,
				MLPService:             mlpSvc,
				ProjectSettingsService: settingsSvc,
				SegmenterService:       segmenterSvc,
			},
		},
	}
}

func TestExperimentController(t *testing.T) {
	suite.Run(t, new(ExperimentControllerTestSuite))
}

func (s *ExperimentControllerTestSuite) TestGetExperiment() {
	t := s.Suite.T()

	tests := []struct {
		name         string
		projectID    int64
		experimentID int64
		expected     string
	}{
		{
			name:         "project settings not found",
			projectID:    1,
			experimentID: 20,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:         "experiment not found",
			projectID:    2,
			experimentID: 20,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"experiment not found\""),
		},
		{
			name:         "mlp project not found",
			projectID:    4,
			experimentID: 2,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:         "success",
			projectID:    2,
			experimentID: 2,
			expected:     fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[0]),
		},
		{
			name:         "success",
			projectID:    5,
			experimentID: 1,
			expected:     fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[2]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetExperiment(w, nil, data.projectID, data.experimentID)
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

func (s *ExperimentControllerTestSuite) TestListExperiments() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		expected  string
	}{
		{
			name:      "project settings not found",
			projectID: 1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "unexpected error",
			projectID: 3,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 500, "\"unexpected error\""),
		},
		{
			name:      "failure | mlp project not found",
			projectID: 4,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 2,
			expected:  fmt.Sprintf(`{"data": [%s]}`, s.expectedExperimentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(nil)))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListExperiments(w, req, data.projectID, api.ListExperimentsParams{})
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

func (s *ExperimentControllerTestSuite) TestCreateExperiment() {
	t := s.Suite.T()

	tests := []struct {
		name           string
		projectID      int64
		experimentData string
		expected       string
	}{
		{
			name:           "failure | missing project settings",
			projectID:      1,
			experimentData: `{"name": "test-exp", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:           "failure | create experiment failed",
			projectID:      2,
			experimentData: `{"name": "test-exp", "updated_by": "test-user"}`,
			expected:       fmt.Sprintf(s.expectedErrorResponseFormat, 500, "\"experiment creation failed\""),
		},
		{
			name:           "failure | mlp project not found",
			projectID:      4,
			experimentData: `{"name": "test-exp", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:           "failure | updated_by cannot be unset",
			projectID:      4,
			experimentData: `{"name": "test-exp"}`,
			expected:       fmt.Sprintf(s.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be unset\""),
		},
		{
			name:           "success | use default tier",
			projectID:      2,
			experimentData: `{"name": "test-exp-2", "updated_by": "test-user"}`,
			expected:       fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[0]),
		},
		{
			name:           "success | use given tier",
			projectID:      2,
			experimentData: `{"name": "test-exp-2", "updated_by": "test-user", "tier": "override"}`,
			expected:       fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[1]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.experimentData)))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.CreateExperiment(w, req, data.projectID)
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

func (s *ExperimentControllerTestSuite) TestUpdateExperiment() {
	t := s.Suite.T()

	tests := []struct {
		name           string
		projectID      int64
		experimentID   int64
		experimentData string
		expected       string
	}{
		{
			name:           "failure | missing project settings",
			projectID:      1,
			experimentID:   1,
			experimentData: `{"name": "test-exp", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:           "failure | update experiment failed",
			projectID:      2,
			experimentID:   1,
			experimentData: `{"segment": { "days_of_week": [1] }, "updated_by": "test-user"}`,
			expected:       fmt.Sprintf(s.expectedErrorResponseFormat, 500, "\"experiment update failed\""),
		},
		{
			name:           "failure | mlp project not found",
			projectID:      4,
			experimentID:   2,
			experimentData: `{"description": "test-description-2", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:           "failure | updated_by cannot be unset",
			projectID:      4,
			experimentID:   2,
			experimentData: `{"description": "test-description-2"}`,
			expected:       fmt.Sprintf(s.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be unset\""),
		},
		{
			name:           "success | use default tier",
			projectID:      2,
			experimentID:   1,
			experimentData: `{"description": "test-description-2", "updated_by": "test-user"}`,
			expected:       fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[0]),
		},
		{
			name:           "success | use given tier",
			projectID:      2,
			experimentID:   1,
			experimentData: `{"description": "test-description-2", "updated_by": "test-user", "tier": "override"}`,
			expected:       fmt.Sprintf(`{"data": %s}`, s.expectedExperimentResponses[1]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.experimentData)))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.UpdateExperiment(w, req, data.projectID, data.experimentID)
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

func (s *ExperimentControllerTestSuite) TestEnableExperiment() {
	t := s.Suite.T()

	tests := []struct {
		name         string
		projectID    int64
		experimentID int64
		expected     string
	}{
		{
			name:         "failure | missing project settings",
			projectID:    1,
			experimentID: 1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:         "failure | mlp project not found",
			projectID:    4,
			experimentID: 2,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:         "failure | mlp project not found",
			projectID:    2,
			experimentID: 3,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 400, "\"experiment id 3 is already active\""),
		},
		{
			name:         "success",
			projectID:    2,
			experimentID: 1,
			expected:     fmt.Sprintf(`{"data": %s}`, "null"),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte{}))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.EnableExperiment(w, req, data.projectID, data.experimentID)
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

func (s *ExperimentControllerTestSuite) TestDisableExperiment() {
	t := s.Suite.T()

	tests := []struct {
		name         string
		projectID    int64
		experimentID int64
		expected     string
	}{
		{
			name:         "failure | missing project settings",
			projectID:    1,
			experimentID: 1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:         "failure | mlp project not found",
			projectID:    4,
			experimentID: 2,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 4 not found in the cache\""),
		},
		{
			name:         "failure | mlp project not found",
			projectID:    2,
			experimentID: 3,
			expected:     fmt.Sprintf(s.expectedErrorResponseFormat, 400, "\"experiment id 3 is already inactive\""),
		},
		{
			name:         "success",
			projectID:    2,
			experimentID: 1,
			expected:     fmt.Sprintf(`{"data": %s}`, "null"),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte{}))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.DisableExperiment(w, req, data.projectID, data.experimentID)
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
