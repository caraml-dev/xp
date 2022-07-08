package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/mlp/api/client"
	"github.com/gojek/xp/common/api/schema"
	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type ProjectSettingsControllerTestSuite struct {
	suite.Suite
	ctrl                                  *ProjectSettingsController
	expectedProjectSettingsResponse       string
	expectedProjectSettingsParamsResponse string
	expectedErrorResponseFormat           string
}

func (s *ProjectSettingsControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ProjectSettingsControllerTestSuite")

	// Create mock project settings service and set up with test responses
	projects := []models.Project{{Id: 1, Segmenters: []string{"test-seg"}}}
	projectSettings := models.Settings{
		ProjectID: 2,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp_var_1", "exp_var_2"},
				},
			},
			RandomizationKey: "rand",
		},
		TreatmentSchema: &models.TreatmentSchema{
			Rules: []models.Rule{
				{
					Name:      "rule_1",
					Predicate: "predicate_1",
				},
				{
					Name:      "rule_2",
					Predicate: "predicate_2",
				},
			},
		},
		ValidationUrl: nil,
	}
	s.expectedProjectSettingsResponse = `{
		"project_id": 2,
		"created_at": "0001-01-01T00:00:00Z",
		"updated_at": "0001-01-01T00:00:00Z",
		"username": "",
		"passkey": "",
		"segmenters": {
			"names": ["seg1"],
			"variables": {
			  "seg1": ["exp_var_1", "exp_var_2"]
			}
		},
		"treatment_schema": {
			"rules": [
				{
					"name": "rule_1",
					"predicate": "predicate_1"
				},
				{
					"name": "rule_2",
					"predicate": "predicate_2"
				}
			]
		},
		"randomization_key": "rand",
		"enable_s2id_clustering": false
	}`
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedProjectSettingsParamsResponse = `{"data": ["rand", "exp_var_1", "exp_var_2"]}`
	segmenter := []*_segmenters.SegmenterConfiguration{
		{
			Constraints: nil,
			MultiValued: false,
			Name:        "test-segmenter",
			Options:     make(map[string]*_segmenters.SegmenterValue),
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"test-segmenter"},
					},
				},
			},
			Type: _segmenters.SegmenterValueType_STRING,
		},
	}
	settingsSvc := &mocks.ProjectSettingsService{}
	settingsSvc.
		On("ListProjects").
		Return(&projects, nil)
	settingsSvc.
		On("GetProjectSettings", int64(1)).
		Return(nil, errors.Newf(errors.NotFound, "test get project settings error"))
	settingsSvc.
		On("GetProjectSettings", int64(2)).
		Return(&projectSettings, nil)
	settingsSvc.
		On("GetExperimentVariables", int64(1)).
		Return(nil, errors.Newf(errors.NotFound, "test get project settings error"))
	settingsSvc.
		On("GetExperimentVariables", int64(2)).
		Return(&[]string{"rand", "exp_var_1", "exp_var_2"}, nil)
	settingsSvc.
		On("ListSegmenters", int64(2)).
		Return(segmenter, nil)
	settingsSvc.
		On("ListProjectSettings").
		Return(&[]models.Settings{projectSettings}, nil)
	settingsSvc.
		On("CreateProjectSettings", int64(2), services.CreateProjectSettingsRequestBody{}).
		Return(nil, fmt.Errorf("test create project settings error"))
	settingsSvc.
		On("CreateProjectSettings", int64(4), services.CreateProjectSettingsRequestBody{
			Username: "client-4",
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp_var_1", "exp_var_2"},
				},
			},
			TreatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "rule_1",
						Predicate: "predicate_1",
					},
					{
						Name:      "rule_2",
						Predicate: "predicate_2",
					},
				},
			},
			ValidationUrl: nil,
		}).
		Return(&projectSettings, nil)
	settingsSvc.
		On("UpdateProjectSettings", int64(2), services.UpdateProjectSettingsRequestBody{RandomizationKey: "rkey1", Segmenters: models.ProjectSegmenters{}}).
		Return(nil, fmt.Errorf("test update project settings internal error"))
	settingsSvc.
		On("UpdateProjectSettings", int64(2), services.UpdateProjectSettingsRequestBody{
			RandomizationKey: "rkey2",
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp_var_1", "exp_var_2"},
				},
			},
			TreatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "rule_1",
						Predicate: "predicate_1",
					},
					{
						Name:      "rule_2",
						Predicate: "predicate_2",
					},
				},
			},
			ValidationUrl: nil,
		}).
		Return(&projectSettings, nil)

	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(&client.Project{Name: ""}, nil)
	mlpSvc.On("GetProject", int64(2)).Return(&client.Project{Name: ""}, nil)
	mlpSvc.On(
		"GetProject", int64(3),
	).Return(&client.Project{Name: ""}, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(3)))
	mlpSvc.On("GetProject", int64(4)).Return(&client.Project{Name: "client-4"}, nil)

	// Create test controller
	s.ctrl = &ProjectSettingsController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				MLPService:             mlpSvc,
				ProjectSettingsService: settingsSvc,
			},
		},
	}
}

func TestProjectSettingsController(t *testing.T) {
	suite.Run(t, new(ProjectSettingsControllerTestSuite))
}

func (s *ProjectSettingsControllerTestSuite) TestGetProjectSettings() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		expected  string
	}{
		{
			name:      "project settings not found",
			projectID: 1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"test get project settings error\""),
		},
		{
			name:      "success",
			projectID: 2,
			expected:  fmt.Sprintf(`{"data": %s}`, s.expectedProjectSettingsResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetProjectSettings(w, nil, data.projectID)
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

func (s *ProjectSettingsControllerTestSuite) TestGetProjectExperimentVariables() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		expected  string
	}{
		{
			name:      "project settings not found",
			projectID: 1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"test get project settings error\""),
		},
		{
			name:      "mlp project not found",
			projectID: 3,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 2,
			expected:  s.expectedProjectSettingsParamsResponse,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetProjectExperimentVariables(w, nil, data.projectID)
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

func (s *ProjectSettingsControllerTestSuite) TestListProjects() {
	w := httptest.NewRecorder()
	s.ctrl.ListProjects(w, nil)
	resp := w.Result()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().JSONEq(`{"data": [{
		"id": 1,
		"username": "",
		"randomization_key": "",
		"segmenters": ["test-seg"],
		"created_at": "0001-01-01T00:00:00Z",
		"updated_at": "0001-01-01T00:00:00Z"
	}]}`, string(body))
}

func (s *ProjectSettingsControllerTestSuite) TestCreateProjectSettings() {
	t := s.Suite.T()

	// Make test requests
	req1, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{}`)))
	s.Suite.Require().NoError(err)
	req2, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`
		{
			"username": "client-4",
			"segmenters": {
				"names": ["seg1"],
				"variables": {
				  "seg1": ["exp_var_1", "exp_var_2"]
				}
			},
			"treatment_schema": {
				"rules": [
					{
						"name": "rule_1",
						"predicate": "predicate_1"
					},
					{
						"name": "rule_2",
						"predicate": "predicate_2"
					}
				]
			}
		}`)))
	s.Suite.Require().NoError(err)
	req3, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(
		[]byte(
			`{	"username": "client-2", 
				"segmenters": {
					"names": ["seg1"],
					"variables": {
					  "seg1": ["exp_var_1", "exp_var_2"]
					}
				}, 
				"treatment_schema": {
					"rules": [
						{
							"name": "rule_1",
							"predicate": "predicate_1"
						},
						{
							"name": "rule_2",
							"predicate": "predicate_2"
						}
					]
				},
				"randomization_key": "random"}`,
		),
	),
	)
	s.Suite.Require().NoError(err)

	tests := []struct {
		name      string
		projectID int64
		request   *http.Request
		expected  string
	}{
		{
			name:      "failure",
			projectID: 2,
			request:   req1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 500, "\"test create project settings error\""),
		},
		{
			name:      "mlp project not found",
			projectID: 3,
			request:   req3,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 4,
			request:   req2,
			expected:  fmt.Sprintf(`{"data": %s}`, s.expectedProjectSettingsResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.CreateProjectSettings(w, data.request, data.projectID)
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

func (s *ProjectSettingsControllerTestSuite) TestUpdateProjectSettings() {
	t := s.Suite.T()

	// Make test requests
	req1, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{}`)))
	s.Suite.Require().NoError(err)
	req2, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"randomization_key": "rkey1"}`)))
	s.Suite.Require().NoError(err)
	req3, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`
		{
			"randomization_key": "rkey2",
			"segmenters": {
				"names": ["seg1"],
				"variables": {
				  "seg1": ["exp_var_1", "exp_var_2"]
				}
			},
			"treatment_schema": {
				"rules": [
					{
						"name": "rule_1",
						"predicate": "predicate_1"
					},
					{
						"name": "rule_2",
						"predicate": "predicate_2"
					}
				]
			}
		}`)))
	s.Suite.Require().NoError(err)
	req4, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(
		[]byte(
			`{"username": "client-2", "randomization_key": "random", 
				"segmenters": {
					"names": ["seg1"],
					"variables": {
					  "seg1": ["seg1"]
					}
				}}`,
		),
	),
	)
	s.Suite.Require().NoError(err)

	tests := []struct {
		name      string
		projectID int64
		request   *http.Request
		expected  string
	}{
		{
			name:      "failure",
			projectID: 1,
			request:   req1,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"test get project settings error\""),
		},
		{
			name:      "failure",
			projectID: 2,
			request:   req2,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 500, "\"test update project settings internal error\""),
		},
		{
			name:      "failure | mlp project not found",
			projectID: 3,
			request:   req4,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
		{
			name:      "success",
			projectID: 2,
			request:   req3,
			expected:  fmt.Sprintf(`{"data": %s}`, s.expectedProjectSettingsResponse),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.UpdateProjectSettings(w, data.request, data.projectID)
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

func (s *ProjectSettingsControllerTestSuite) TestParseTreatmentSchema() {
	tests := []struct {
		treatmentSchema *schema.TreatmentSchema
		expected        *models.TreatmentSchema
	}{
		{
			treatmentSchema: nil,
			expected:        nil,
		},
		{
			treatmentSchema: &schema.TreatmentSchema{
				Rules: []schema.Rule{
					{
						Name:      "rule-1",
						Predicate: "{{- ( .field1) -}}",
					},
					{
						Name:      "rule-2",
						Predicate: "{{- ( .field2) -}}",
					},
				},
			},
			expected: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "rule-1",
						Predicate: "{{- ( .field1) -}}",
					},
					{
						Name:      "rule-2",
						Predicate: "{{- ( .field2) -}}",
					},
				},
			},
		},
	}

	for i, data := range tests {
		s.Suite.T().Run(fmt.Sprintf("Test %v", i), func(t *testing.T) {
			actual := parseTreatmentSchema(data.treatmentSchema)
			s.Suite.Assert().Equal(data.expected, actual)
		})
	}
}
