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

	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type TreatmentHistoryControllerTestSuite struct {
	suite.Suite
	ctrl                             *TreatmentHistoryController
	expectedTreatmentHistoryResponse string
	expectedErrorResponseFormat      string
}

func (s *TreatmentHistoryControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up TreatmentHistoryControllerTestSuite")

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

	// Create mock treatment service and set up with test responses
	treatmentSvc := &mocks.TreatmentService{}
	treatmentSvc.
		On("GetDBRecord", models.ID(3), models.ID(1)).
		Return(nil, errors.Newf(errors.NotFound, "treatment not found"))
	treatmentSvc.
		On("GetDBRecord", models.ID(3), models.ID(10)).
		Return(nil, nil)

	// Set up mock treatment history service
	var testConfig = map[string]interface{}{
		"config-1": "value",
		"config-2": 2,
	}
	testTreatmentHistory := &models.TreatmentHistory{
		Model: models.Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:            models.ID(100),
		TreatmentID:   models.ID(40),
		Version:       int64(8),
		Name:          "treatment-hist-1",
		Configuration: testConfig,
		UpdatedBy:     "test-updated-by",
	}
	treatmentHistSvc := &mocks.TreatmentHistoryService{}
	treatmentHistSvc.
		On("GetTreatmentHistory", int64(10), int64(2)).
		Return(nil, errors.Newf(errors.NotFound, "treatment history not found"))
	treatmentHistSvc.
		On("GetTreatmentHistory", int64(10), int64(1)).
		Return(testTreatmentHistory, nil)
	treatmentHistSvc.
		On("ListTreatmentHistory", int64(10), services.ListTreatmentHistoryParams{
			PaginationOptions: pagination.PaginationOptions{},
		}).
		Return([]*models.TreatmentHistory{testTreatmentHistory}, &pagination.Paging{Page: 1, Total: 1, Pages: 1}, nil)

	// Set up expected responses
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedTreatmentHistoryResponse = `{
		"created_at": "2021-01-01T02:03:04Z",
		"updated_at": "2021-01-01T02:03:04Z",
		"treatment_id": 40,
		"id": 100,
		"name": "treatment-hist-1",
		"configuration": {
			"config-1": "value",
			"config-2": 2
		},
		"updated_by": "test-updated-by",
		"version": 8
	}`

	// Create test controller
	s.ctrl = &TreatmentHistoryController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				TreatmentService:        treatmentSvc,
				TreatmentHistoryService: treatmentHistSvc,
				MLPService:              mlpSvc,
				ProjectSettingsService:  settingsSvc,
			},
		},
	}
}

func TestTreatmentHistoryController(t *testing.T) {
	suite.Run(t, new(TreatmentHistoryControllerTestSuite))
}

func (s *TreatmentHistoryControllerTestSuite) TestGetTreatmentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		treatmentID int64
		version     int64
		expected    string
	}{
		{
			name:        "mlp project not found",
			projectID:   1,
			treatmentID: 1,
			version:     1,
			expected:    fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:        "project settings not found",
			projectID:   2,
			treatmentID: 1,
			version:     1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:        "treatment not found",
			projectID:   3,
			treatmentID: 1,
			version:     1,
			expected:    fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Treatment with id 1 cannot be retrieved: treatment not found\""),
		},
		{
			name:        "treatment history not found",
			projectID:   3,
			treatmentID: 10,
			version:     2,
			expected:    fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"treatment history not found\""),
		},
		{
			name:        "success",
			projectID:   3,
			treatmentID: 10,
			version:     1,
			expected:    fmt.Sprintf(`{"data": %s}`, s.expectedTreatmentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetTreatmentHistory(w, nil, data.projectID, data.treatmentID, data.version)
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

func (s *TreatmentHistoryControllerTestSuite) TestListTreatmentHistory() {
	t := s.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		treatmentID int64
		params      api.ListTreatmentHistoryParams
		expected    string
	}{
		{
			name:        "mlp project not found",
			projectID:   1,
			treatmentID: 1,
			expected:    fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 1 not found in the cache\""),
		},
		{
			name:        "project settings not found",
			projectID:   2,
			treatmentID: 1,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 2 cannot be retrieved: test get project settings error\""),
		},
		{
			name:        "treatment not found",
			projectID:   3,
			treatmentID: 1,
			expected:    fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"Treatment with id 1 cannot be retrieved: treatment not found\""),
		},
		{
			name:        "success",
			projectID:   3,
			treatmentID: 10,
			expected:    fmt.Sprintf(`{"data": [%s], "paging": {"page": 1, "pages": 1, "total": 1}}`, s.expectedTreatmentHistoryResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(nil))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListTreatmentHistory(w, req, data.projectID, data.treatmentID, data.params)
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
