package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type TreatmentControllerTestSuite struct {
	suite.Suite
	ctrl                        *TreatmentController
	expectedTreatmentResponses  []string
	expectedErrorResponseFormat string
}

func (s *TreatmentControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up TreatmentControllerTestSuite")

	// Configure expected responses and errors
	s.expectedTreatmentResponses = []string{
		`{
			"project_id": 2,
			"created_at": "0001-01-01T00:00:00Z",
			"id": 0,
			"name": "",
			"configuration": {"team": "business"},
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
			Config: &models.ExperimentationConfig{Segmenters: models.ProjectSegmenters{
				Names:     []string{""},
				Variables: nil,
			}},
		}, nil)
	settingsSvc.
		On("GetProjectSettings", int64(4)).
		Return(nil, nil)

	// Create mock treatment service and set up with test responses
	treatmentSvc := &mocks.TreatmentService{}
	testTreatment := &models.Treatment{ProjectID: 2, Configuration: map[string]interface{}{"team": "business"}}
	treatmentSvc.
		On("GetTreatment", int64(2), int64(20)).
		Return(nil, errors.Newf(errors.NotFound, "treatment not found"))
	treatmentSvc.
		On("GetTreatment", int64(2), int64(2)).
		Return(testTreatment, nil)

	treatmentSvc.
		On("ListTreatments", int64(2), services.ListTreatmentsParams{}).
		Return([]*models.Treatment{testTreatment}, nil, nil)
	treatmentSvc.
		On("ListTreatments", int64(4), services.ListTreatmentsParams{}).
		Return(nil, nil, fmt.Errorf("unexpected error"))
	updatedBy := "test-user"
	treatmentSvc.
		On("CreateTreatment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateTreatmentRequestBody{Name: "test-treatment", UpdatedBy: &updatedBy}).
		Return(nil, fmt.Errorf("treatment creation failed"))
	treatmentSvc.
		On("CreateTreatment",
			models.Settings{ProjectID: models.ID(2)},
			services.CreateTreatmentRequestBody{Name: "test-treatment-2", Config: map[string]interface{}{"team": "business"}, UpdatedBy: &updatedBy}).
		Return(testTreatment, nil)
	treatmentSvc.
		On("UpdateTreatment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateTreatmentRequestBody{Config: nil, UpdatedBy: &updatedBy}).
		Return(nil, fmt.Errorf("treatment update failed"))
	treatmentSvc.
		On("UpdateTreatment",
			models.Settings{ProjectID: models.ID(2)},
			int64(1),
			services.UpdateTreatmentRequestBody{Config: map[string]interface{}{"team": "business"}, UpdatedBy: &updatedBy}).
		Return(testTreatment, nil)
	treatmentSvc.
		On("DeleteTreatment",
			int64(2), int64(2)).
		Return(nil)

	// Create mock MLP service and set up with test responses
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(2)).Return(nil, nil)
	mlpSvc.On(
		"GetProject", int64(3),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(3)))
	mlpSvc.On("GetProject", int64(4)).Return(nil, nil)

	// Create test controller
	s.ctrl = &TreatmentController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				TreatmentService:       treatmentSvc,
				MLPService:             mlpSvc,
				ProjectSettingsService: settingsSvc,
			},
		},
	}
}

func TestTreatmentController(t *testing.T) {
	suite.Run(t, new(TreatmentControllerTestSuite))
}

func (p *TreatmentControllerTestSuite) TestGetTreatment() {
	t := p.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		treatmentID int64
		expected    string
	}{
		{
			name:        "failure | project settings not found",
			projectID:   1,
			treatmentID: 20,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:        "failure | treatment not found",
			projectID:   2,
			treatmentID: 20,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"treatment not found\""),
		},
		{
			name:        "failure | mlp project not found",
			projectID:   3,
			treatmentID: 2,
			expected:    fmt.Sprintf(p.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:        "success",
			projectID:   2,
			treatmentID: 2,
			expected:    fmt.Sprintf(`{"data": %s}`, p.expectedTreatmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.GetTreatment(w, nil, data.projectID, data.treatmentID)
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

func (p *TreatmentControllerTestSuite) TestListTreatments() {
	t := p.Suite.T()

	tests := []struct {
		name      string
		projectID int64
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
			name:      "success",
			projectID: 2,
			expected:  fmt.Sprintf(`{"data": [%s]}`, p.expectedTreatmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(nil)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.ListTreatments(w, req, data.projectID, api.ListTreatmentsParams{})
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

func (p *TreatmentControllerTestSuite) TestCreateTreatment() {
	t := p.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		treatmentData string
		expected      string
	}{
		{
			name:          "failure | missing project settings",
			projectID:     1,
			treatmentData: `{"name": "test-treatment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:          "failure | create treatment failed",
			projectID:     2,
			treatmentData: `{"name": "test-treatment", "updated_by": "test-user"}`,
			expected:      fmt.Sprintf(p.expectedErrorResponseFormat, 500, "\"treatment creation failed\""),
		},
		{
			name:          "failure | mlp project not found",
			projectID:     3,
			treatmentData: `{"name": "test-treatment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:          "failure | updated_by cannot be empty",
			projectID:     4,
			treatmentData: `{"name": "test-exp"}`,
			expected:      fmt.Sprintf(p.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be empty\""),
		},
		{
			name:          "success",
			projectID:     2,
			treatmentData: `{"name": "test-treatment-2", "configuration": {"team": "business"}, "updated_by": "test-user"}`,
			expected:      fmt.Sprintf(`{"data": %s}`, p.expectedTreatmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.treatmentData)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.CreateTreatment(w, req, data.projectID)
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

func (p *TreatmentControllerTestSuite) TestUpdateTreatment() {
	t := p.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		treatmentID   int64
		treatmentData string
		expected      string
	}{
		{
			name:          "failure | missing project settings",
			projectID:     1,
			treatmentID:   1,
			treatmentData: `{"name": "test-treatment", "updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test find project settings error\""),
		},
		{
			name:          "failure | update treatment failed",
			projectID:     2,
			treatmentID:   1,
			treatmentData: `{"config": null, "updated_by": "test-user"}`,
			expected:      fmt.Sprintf(p.expectedErrorResponseFormat, 500, "\"treatment update failed\""),
		},
		{
			name:          "failure | mlp project not found",
			projectID:     3,
			treatmentID:   2,
			treatmentData: `{"updated_by": "test-user"}`,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:          "failure | updated_by cannot be empty",
			projectID:     4,
			treatmentID:   2,
			treatmentData: `{}`,
			expected:      fmt.Sprintf(p.expectedErrorResponseFormat, 400, "\"field (updated_by) cannot be empty\""),
		},
		{
			name:          "success",
			projectID:     2,
			treatmentID:   1,
			treatmentData: `{"configuration": {"team": "business"}, "updated_by": "test-user"}`,
			expected:      fmt.Sprintf(`{"data": %s}`, p.expectedTreatmentResponses[0]),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.treatmentData)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.UpdateTreatment(w, req, data.projectID, data.treatmentID)
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

func (p *TreatmentControllerTestSuite) TestDeleteTreatment() {
	t := p.Suite.T()

	tests := []struct {
		name        string
		projectID   int64
		treatmentID int64
		expected    string
	}{
		{
			name:        "failure | missing project settings",
			projectID:   1,
			treatmentID: 1,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"Settings for project_id 1 cannot be retrieved: test get project settings error\""),
		},
		{
			name:        "failure | mlp project not found",
			projectID:   3,
			treatmentID: 2,
			expected: fmt.Sprintf(p.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:        "success",
			projectID:   2,
			treatmentID: 2,
			expected:    fmt.Sprintf(`{"data": %s}`, p.expectedTreatmentResponses[1]),
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
			p.ctrl.DeleteTreatment(w, req, data.projectID, data.treatmentID)
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
