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

type SegmentHistoryControllerTestSuite struct {
	suite.Suite
	ctrl                           *SegmentHistoryController
	expectedSegmentHistoryResponse string
	expectedErrorResponseFormat    string
}

func (s *SegmentHistoryControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmentHistoryControllerTestSuite")

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

	// Create mock segment service and set up with test responses
	segmentSvc := &mocks.SegmentService{}
	segmentSvc.
		On("GetDBRecord", models.ID(3), models.ID(1)).
		Return(nil, errors.Newf(errors.NotFound, "segment not found"))
	segmentSvc.
		On("GetDBRecord", models.ID(3), models.ID(10)).
		Return(nil, nil)

	// Set up mock segment history service
	hoursOfDay := []string{"17", "18", "19", "20"}
	daysOfWeek := []string{"1", "2"}
	experimentSegment := models.ExperimentSegment{
		"days_of_week": daysOfWeek,
		"hours_of_day": hoursOfDay,
	}

	testSegmentHistory := &models.SegmentHistory{
		Model: models.Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:        models.ID(100),
		SegmentID: models.ID(40),
		Version:   int64(8),
		Name:      "segment-hist-1",
		Segment:   experimentSegment,
		UpdatedBy: "test-updated-by",
	}
	segmentHistSvc := &mocks.SegmentHistoryService{}
	segmentHistSvc.
		On("GetSegmentHistory", int64(10), int64(2)).
		Return(nil, errors.Newf(errors.NotFound, "segment history not found"))
	segmentHistSvc.
		On("GetSegmentHistory", int64(10), int64(1)).
		Return(testSegmentHistory, nil)
	segmentHistSvc.
		On("ListSegmentHistory", int64(10), services.ListSegmentHistoryParams{
			PaginationOptions: pagination.PaginationOptions{},
		}).
		Return([]*models.SegmentHistory{testSegmentHistory}, &pagination.Paging{Page: 1, Total: 1, Pages: 1}, nil)

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("GetSegmenterTypes", int64(3)).
		Return(
			map[string]schema.SegmenterType{
				"hours_of_day": schema.SegmenterTypeInteger,
				"days_of_week": schema.SegmenterTypeInteger,
			},
			nil,
		)

	// Set up expected responses
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedSegmentHistoryResponse = `{
		"created_at": "2021-01-01T02:03:04Z",
		"updated_at": "2021-01-01T02:03:04Z",
		"segment_id": 40,
		"id": 100,
		"name": "segment-hist-1",
		"segment": {
			"days_of_week": [1, 2],
			"hours_of_day": [17, 18, 19, 20]
		},
		"updated_by": "test-updated-by",
		"version": 8
	}`

	// Create test controller
	s.ctrl = &SegmentHistoryController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				SegmentService:         segmentSvc,
				SegmentHistoryService:  segmentHistSvc,
				MLPService:             mlpSvc,
				ProjectSettingsService: settingsSvc,
				SegmenterService:       segmenterSvc,
			},
		},
	}
}

func TestSegmentHistoryController(t *testing.T) {
	suite.Run(t, new(SegmentHistoryControllerTestSuite))
}

func (s *SegmentHistoryControllerTestSuite) TestGetSegmentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		segmentID int64
		version   int64
		expected  string
	}{
		{
			name:      "mlp project not found",
			projectID: 1,
			segmentID: 1,
			version:   1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:      "project settings not found",
			projectID: 2,
			segmentID: 1,
			version:   1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "segment not found",
			projectID: 3,
			segmentID: 1,
			version:   1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Segment with id 1 cannot be retrieved: segment not found\""),
		},
		{
			name:      "segment history not found",
			projectID: 3,
			segmentID: 10,
			version:   2,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"segment history not found\""),
		},
		{
			name:      "success",
			projectID: 3,
			segmentID: 10,
			version:   1,
			expected:  fmt.Sprintf(`{"data": %s}`, s.expectedSegmentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetSegmentHistory(w, nil, data.projectID, data.segmentID, data.version)
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

func (s *SegmentHistoryControllerTestSuite) TestListSegmentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		segmentID int64
		params    api.ListSegmentHistoryParams
		expected  string
	}{
		{
			name:      "mlp project not found",
			projectID: 1,
			segmentID: 1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:      "project settings not found",
			projectID: 2,
			segmentID: 1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:      "segment not found",
			projectID: 3,
			segmentID: 1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Segment with id 1 cannot be retrieved: segment not found\""),
		},
		{
			name:      "success",
			projectID: 3,
			segmentID: 10,
			expected:  fmt.Sprintf(`{"data": [%s], "paging": {"page": 1, "pages": 1, "total": 1}}`, s.expectedSegmentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(nil))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListSegmentHistory(w, req, data.projectID, data.segmentID, data.params)
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
