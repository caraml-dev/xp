package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/caraml-dev/xp/common/api/schema"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/mocks"
)

type SegmentControllerTestSuite struct {
	suite.Suite
	ctrl                        *SegmentController
	expectedSegmentResponses    []string
	expectedErrorResponseFormat string
}

func (s *SegmentControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmentControllerTestSuite")

	// Configure expected responses and errors
	s.expectedSegmentResponses = []string{
		`{
			"project_id": 2,
			"created_at": "0001-01-01T00:00:00Z",
			"id": 0,
			"name": "",
			"segment": {"days_of_week": [1]},
			"updated_at": "0001-01-01T00:00:00Z",
			"updated_by": ""
		}`,
		`
		{
			"id": 2
		}
		`,
	}
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`

	// Create mock project settings service and set up with test responses
	daysOfWeekRaw := []interface{}{float64(1)}
	daysOfWeek2Raw := []interface{}{float64(2)}
	daysOfWeek := []string{"1"}
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
		On("GetProjectSettings", int64(4)).
		Return(nil, nil)

	// Create mock segment service and set up with test responses
	segmentSvc := &mocks.SegmentService{}
	testSegment := &models.Segment{ProjectID: 2, Segment: models.ExperimentSegment{"days_of_week": daysOfWeek}}
	segmentSvc.
		On("GetSegment", int64(2), int64(20)).
		Return(nil, errors.Newf(errors.NotFound, "segment not found"))
	segmentSvc.
		On("GetSegment", int64(2), int64(2)).
		Return(testSegment, nil)

	segmentSvc.
		On("ListSegments", int64(2), mock.Anything).
		Return([]*models.Segment{testSegment}, nil, nil)
	segmentSvc.
		On("ListSegments", int64(4), mock.Anything).
		Return(nil, nil, fmt.Errorf("unexpected error"))
	updatedBy := "test-user"
	segmentSvc.
		On("CreateSegment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateSegmentRequestBody{Name: "test-segment", UpdatedBy: &updatedBy, Segment: models.ExperimentSegmentRaw(nil)}).
		Return(nil, fmt.Errorf("segment creation failed"))
	segmentSvc.
		On("CreateSegment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateSegmentRequestBody{
				Name:      "test-segment-2",
				Segment:   models.ExperimentSegmentRaw{"days_of_week": daysOfWeekRaw},
				UpdatedBy: &updatedBy,
			}).
		Return(testSegment, nil)
	segmentSvc.
		On("UpdateSegment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateSegmentRequestBody{
				Segment:   models.ExperimentSegmentRaw(nil),
				UpdatedBy: &updatedBy,
			}).
		Return(nil, fmt.Errorf("segment update failed"))
	segmentSvc.
		On("UpdateSegment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateSegmentRequestBody{
				Segment:   models.ExperimentSegmentRaw{"days_of_week": daysOfWeek2Raw},
				UpdatedBy: &updatedBy,
			}).
		Return(testSegment, nil)
	segmentSvc.
		On("DeleteSegment",
			int64(2), int64(2)).
		Return(nil)

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("ListGlobalSegmentersNames").
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
		On("GetSegmenterTypes", int64(2)).
		Return(
			map[string]schema.SegmenterType{
				"days_of_week": schema.SegmenterTypeInteger,
			},
			nil,
		)

	// Create mock MLP service and set up with test responses
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(2)).Return(nil, nil)
	mlpSvc.On(
		"GetProject", int64(3),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(3)))
	mlpSvc.On("GetProject", int64(4)).Return(nil, nil)

	// Create test controller
	s.ctrl = &SegmentController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				SegmentService:         segmentSvc,
				MLPService:             mlpSvc,
				ProjectSettingsService: settingsSvc,
				SegmenterService:       segmenterSvc,
			},
		},
	}
}

func TestSegmentController(t *testing.T) {
	suite.Run(t, new(SegmentControllerTestSuite))
}

func (p *SegmentControllerTestSuite) TestGetSegment() {
	t := p.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		segmentID int64
		expected  string
	}{
		{
			name:      "failure | project settings not found",
			projectID: 1,
			segmentID: 20,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "failure | segment not found",
			projectID: 2,
			segmentID: 20,
			expected:  fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"segment not found\""),
		},
		{
			name:      "failure | mlp project not found",
			projectID: 3,
			segmentID: 2,
			expected:  fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 2,
			segmentID: 2,
			expected:  fmt.Sprintf(`{"data": %s}`, p.expectedSegmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.GetSegment(w, nil, data.projectID, data.segmentID)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			p.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (p *SegmentControllerTestSuite) TestListSegments() {
	t := p.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		params    api.ListSegmentsParams
		expected  string
	}{
		{
			name:      "failure | project settings not found",
			projectID: 1,
			expected:  fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "failure | mlp project not found",
			projectID: 3,
			expected:  fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "failure | unexpected error",
			projectID: 4,
			expected:  fmt.Sprintf(p.expectedErrorResponseFormat, 500, "\"unexpected error\""),
		},
		{
			name:      "success | all fields",
			projectID: 2,
			expected:  fmt.Sprintf(`{"data": [%s]}`, p.expectedSegmentResponses[0]),
		},
		{
			name:      "success | fields filter",
			projectID: 2,
			params:    api.ListSegmentsParams{Fields: &[]schema.SegmentField{schema.SegmentFieldId}},
			expected:  `{"data": [{"id": 0}]}`,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(nil)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.ListSegments(w, req, data.projectID, data.params)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			p.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (p *SegmentControllerTestSuite) TestCreateSegment() {
	t := p.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		segmentData string
		expected    string
	}{
		{
			name:        "failure | missing project settings",
			projectID:   1,
			segmentData: `{"name": "test-segment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:        "failure | create segment failed",
			projectID:   2,
			segmentData: `{"name": "test-segment", "updated_by": "test-user"}`,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 500, "\"segment creation failed\""),
		},
		{
			name:        "failure | mlp project not found",
			projectID:   3,
			segmentData: `{"name": "test-segment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:        "failure | updated_by cannot be empty",
			projectID:   4,
			segmentData: `{"name": "test-exp"}`,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be empty\""),
		},
		{
			name:        "success",
			projectID:   2,
			segmentData: `{"name": "test-segment-2", "segment": {"days_of_week": [1]}, "updated_by": "test-user"}`,
			expected:    fmt.Sprintf(`{"data": %s}`, p.expectedSegmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.segmentData)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.CreateSegment(w, req, data.projectID)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			p.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (p *SegmentControllerTestSuite) TestUpdateSegment() {
	t := p.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		segmentID   int64
		segmentData string
		expected    string
	}{
		{
			name:        "failure | missing project settings",
			projectID:   1,
			segmentID:   1,
			segmentData: `{"name": "test-segment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:        "failure | update segment failed",
			projectID:   2,
			segmentID:   1,
			segmentData: `{"updated_by": "test-user"}`,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 500, "\"segment update failed\""),
		},
		{
			name:        "failure | mlp project not found",
			projectID:   3,
			segmentID:   2,
			segmentData: `{"updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:        "failure | updated_by cannot be empty",
			projectID:   4,
			segmentID:   2,
			segmentData: `{}`,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be empty\""),
		},
		{
			name:        "success",
			projectID:   2,
			segmentID:   1,
			segmentData: `{"segment": {"days_of_week": [2]}, "updated_by": "test-user"}`,
			expected:    fmt.Sprintf(`{"data": %s}`, p.expectedSegmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.segmentData)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.UpdateSegment(w, req, data.projectID, data.segmentID)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			p.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (p *SegmentControllerTestSuite) TestDeleteSegment() {
	t := p.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		segmentID int64
		expected  string
	}{
		{
			name:      "failure | missing project settings",
			projectID: 1,
			segmentID: 1,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "failure | mlp project not found",
			projectID: 3,
			segmentID: 2,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 2,
			segmentID: 2,
			expected:  fmt.Sprintf(`{"data": %s}`, p.expectedSegmentResponses[1]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(nil)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.DeleteSegment(w, req, data.projectID, data.segmentID)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			p.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}
