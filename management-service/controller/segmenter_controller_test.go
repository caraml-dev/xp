package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/segmenters"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type SegmenterControllerTestSuite struct {
	suite.Suite
	ctrl                        *SegmenterController
	expectedErrorResponseFormat string
	expectedSegmentersResponses []string
}

func (s *SegmenterControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmenterControllerTestSuite")

	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedSegmentersResponses = []string{
		`{
			"data": [{
				"constraints": null,
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-global-segmenter",
				"options": {},
				"treatment_request_fields": [
						["test-global-segmenter"]
					],
				"type": "string",
				"required": false
			}]
		}`,
		`{
			"data": [{
				"constraints": null,
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-global-segmenter",
				"options": {},
				"treatment_request_fields": [
						["test-global-segmenter"]
					],
				"type": "string",
				"required": false,
				"scope": "global", 
				"status": "active"
			}]
		}`,
		`{
			"data": [{
				"constraints": null,
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-global-segmenter",
				"options": {},
				"treatment_request_fields": [
						["test-global-segmenter"]
					],
				"type": "string",
				"required": false,
				"scope": "global", 
				"status": "inactive"
			}]
		}`,
		`{
			"data": [
				{
					"constraints": null,
					"description": "Test Desc",
					"multi_valued": false,
					"name": "test-global-segmenter",
					"options": {},
					"treatment_request_fields": [
							["test-global-segmenter"]
						],
					"type": "string",
					"required": false,
					"scope": "global", 
					"status": "inactive"
				},
				{
					"constraints": null,
					"description": "Test Desc",
					"multi_valued": false,
					"name": "test-custom-segmenter",
					"options": {},
					"treatment_request_fields": [["test-custom-segmenter"]],
					"type":"string",
					"required": false,
					"scope": "project",
					"status": "active",
					"created_at": "0001-01-01T00:00:00Z", 
					"updated_at": "0001-01-01T00:00:00Z"
				}
			]
		}`,
		`{
			"data": [
				{
					"constraints": null,
					"description": "Test Desc",
					"multi_valued": false,
					"name": "test-global-segmenter",
					"options": {},
					"treatment_request_fields": [
							["test-global-segmenter"]
						],
					"type": "string",
					"required": false,
					"scope": "global", 
					"status": "inactive"
				},
				{
					"constraints": null, 
					"description": "Test Desc",
					"multi_valued": false,
					"name": "test-custom-segmenter",
					"options": {},
					"treatment_request_fields": [["test-custom-segmenter"]],
					"type":"string",
					"required": false,
					"scope": "project",
					"status": "inactive",
					"created_at": "0001-01-01T00:00:00Z", 
					"updated_at": "0001-01-01T00:00:00Z"
				}
			]
		}`,
		`{
			"data": [
				{
					"constraints": null, 
					"description": "Test Desc",
					"multi_valued": false,
					"name": "test-custom-segmenter",
					"options": {},
					"treatment_request_fields": [["test-custom-segmenter"]],
					"type":"string",
					"required": false,
					"scope": "project",
					"status": "active",
					"created_at": "0001-01-01T00:00:00Z", 
					"updated_at": "0001-01-01T00:00:00Z"
				}
			]
		}`,
		`{
			"data": []
		}`,
		`{
			"data": {
				"constraints": null,
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-global-segmenter",
				"options": {},
				"scope": "global",
				"status": "active",
				"treatment_request_fields": [
						["test-global-segmenter"]
					],
				"type": "string",
				"required": false
			}
		}`,
		`{
			"data": {
				"constraints": null, 
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-custom-segmenter",
				"options": {},
				"scope": "project",
				"status": "inactive",
				"treatment_request_fields": [["test-custom-segmenter"]],
				"type":"string",
				"required": false,
				"created_at": "0001-01-01T00:00:00Z", 
				"updated_at": "0001-01-01T00:00:00Z"
			}
		}`,
		`{
			"data": {
				"name": "test-custom-segmenter"
			}
		}`,
		`{
			"data": {
				"constraints": [
					{
						"allowed_values": [
							1
						],
						"options": {
							"option_1": 1
						},
						"pre_requisites": [
							{
								"segmenter_name": "segmenter_name_1",
								"segmenter_values": [
									"SN1"
								]
							}
						]
					}
				],
				"description": "Test Desc",
				"multi_valued": false,
				"name": "test-new-custom-segmenter", 
				"options": {
					"option_a": true
				},
				"required":	false,
				"treatment_request_fields": [["test-new-custom-segmenter"]],
				"type": "string",
				"created_at": "0001-01-01T00:00:00Z",
				"updated_at": "0001-01-01T00:00:00Z"
			}
		}`,
		`{
			"data": {
				"constraints": [
					{
						"allowed_values": [
							1
						],
						"options": {
							"option_1": 1
						},
						"pre_requisites": [
							{
								"segmenter_name": "segmenter_name_1",
								"segmenter_values": [
									"SN1"
								]
							}
						]
					}
				],
				"description": "updated", 
				"multi_valued": false,
				"name": "test-custom-segmenter",
				"options": {
					"option_a": true
				},
				"required": true,
				"treatment_request_fields": [["test-custom-segmenter"]],
				"type": "string",
				"created_at": "0001-01-01T00:00:00Z",
				"updated_at":"0001-01-01T00:00:00Z"
			}
		}`,
		`{
			"data":[
				{
					"constraints":null,
					"description":"Test Desc",
					"multi_valued":false,
					"name":"test-global-segmenter",
					"options":{},
					"required":false,
					"scope":"global",
					"status":"active",
					"treatment_request_fields":[["test-global-segmenter"]],
					"type":"string"
				}
			]
		}`,
	}

	segmentersDescription := "Test Desc"
	baseGlobalSegmenter := segmenters.Segmenter(
		segmenters.NewBaseSegmenter(
			&_segmenters.SegmenterConfiguration{
				Constraints: nil,
				MultiValued: false,
				Description: segmentersDescription,
				Name:        "test-global-segmenter",
				Options:     make(map[string]*_segmenters.SegmenterValue),
				TreatmentRequestFields: &_segmenters.ListExperimentVariables{
					Values: []*_segmenters.ExperimentVariables{
						{
							Value: []string{"test-global-segmenter"},
						},
					},
				},
				Type: _segmenters.SegmenterValueType_STRING,
			},
		),
	)
	baseGlobalSegmenters := []*segmenters.Segmenter{
		&baseGlobalSegmenter,
	}
	// Create variants of baseGlobalSegmenterOpenApi with respect to their scope and status combinations
	config, err := baseGlobalSegmenter.GetConfiguration()
	s.Suite.Require().NoError(err)
	formattedGlobalSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(config)
	s.Suite.Require().NoError(err)

	segmenterScopeGlobal := schema.SegmenterScopeGlobal
	segmenterScopeProject := schema.SegmenterScopeProject
	segmenterStatusActive := schema.SegmenterStatusActive
	segmenterStatusInactive := schema.SegmenterStatusInactive

	activeGlobalSegmenterOpenApi := *formattedGlobalSegmenter
	activeGlobalSegmenterOpenApi.Scope = &segmenterScopeGlobal
	activeGlobalSegmenterOpenApi.Status = &segmenterStatusActive

	inactiveGlobalSegmenterOpenApi := *formattedGlobalSegmenter
	inactiveGlobalSegmenterOpenApi.Scope = &segmenterScopeGlobal
	inactiveGlobalSegmenterOpenApi.Status = &segmenterStatusInactive

	baseCustomSegmenter := models.CustomSegmenter{
		Name:        "test-custom-segmenter",
		ProjectID:   4,
		Type:        models.SegmenterValueTypeString,
		Description: &segmentersDescription,
	}

	// Create variants of baseCustomSegmenterOpenApi with respect to their scope and status combinations
	baseCustomSegmenterOpenApi := baseCustomSegmenter.ToApiSchema()
	activeCustomSegmenterOpenApi := baseCustomSegmenterOpenApi
	activeCustomSegmenterOpenApi.Scope = &segmenterScopeProject
	activeCustomSegmenterOpenApi.Status = &segmenterStatusActive

	inactiveCustomSegmenterOpenApi := baseCustomSegmenterOpenApi
	inactiveCustomSegmenterOpenApi.Scope = &segmenterScopeProject
	inactiveCustomSegmenterOpenApi.Status = &segmenterStatusInactive

	newCustomSegmenter := models.CustomSegmenter{
		Name:        "test-new-custom-segmenter",
		ProjectID:   4,
		Type:        models.SegmenterValueTypeString,
		Description: &segmentersDescription,
		Options: &models.Options{
			"option_a": true,
		},
		Constraints: &models.Constraints{
			{
				PreRequisites: []models.PreRequisite{
					{
						SegmenterName: "segmenter_name_1",
						SegmenterValues: []interface{}{
							"SN1",
						},
					},
				},
				AllowedValues: []interface{}{
					float64(1),
				},
				Options: &models.Options{
					"option_1": float64(1),
				},
			},
		},
	}

	updatedDescription := "updated"
	updatedCustomSegmenter := models.CustomSegmenter{
		Name:        "test-custom-segmenter",
		ProjectID:   4,
		Type:        models.SegmenterValueTypeString,
		Required:    true,
		Description: &updatedDescription,
		Options: &models.Options{
			"option_a": true,
		},
		Constraints: &models.Constraints{
			{
				PreRequisites: []models.PreRequisite{
					{
						SegmenterName: "segmenter_name_1",
						SegmenterValues: []interface{}{
							"SN1",
						},
					},
				},
				AllowedValues: []interface{}{
					float64(1),
				},
				Options: &models.Options{
					"option_1": float64(1),
				},
			},
		},
	}

	projectSettings3 := models.Settings{
		ProjectID: 3,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"test-global-segmenter"},
				Variables: map[string][]string{
					"test-global-segmenter": {"exp-var-test"},
				},
			},
			RandomizationKey: "rand",
		},
	}
	projectSettings4 := models.Settings{
		ProjectID: 4,
		Config: &models.ExperimentationConfig{
			RandomizationKey: "rand",
		},
	}
	projectSettings5 := models.Settings{
		ProjectID: 5,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"test-custom-segmenter"},
				Variables: map[string][]string{
					"test-custom-segmenter": {"exp-var-test"},
				},
			},
			RandomizationKey: "rand",
		},
	}
	projectSettings6 := models.Settings{
		ProjectID: 6,
		Config: &models.ExperimentationConfig{
			RandomizationKey: "rand",
		},
	}
	projectSettings7 := models.Settings{
		ProjectID: 7,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"test-global-segmenter", "test-custom-segmenter"},
				Variables: map[string][]string{
					"test-global-segmenter": {"exp-var-test"},
					"test-custom-segmenter": {"exp-var-test"},
				},
			},
			RandomizationKey: "rand",
		},
	}

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("getGlobalSegmenters").
		Return(baseGlobalSegmenters, nil)

	segmenterSvc.
		On("ListGlobalSegmenters").
		Return([]*schema.Segmenter{&activeGlobalSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(3),
			services.ListSegmentersParams{
				Scope:  nil,
				Status: nil,
			}).
		Return([]*schema.Segmenter{&activeGlobalSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(4),
			services.ListSegmentersParams{
				Scope:  nil,
				Status: nil,
			}).
		Return([]*schema.Segmenter{&inactiveGlobalSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(5),
			services.ListSegmentersParams{
				Scope:  nil,
				Status: nil,
			}).
		Return([]*schema.Segmenter{&inactiveGlobalSegmenterOpenApi, &activeCustomSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(6),
			services.ListSegmentersParams{
				Scope:  nil,
				Status: nil,
			}).
		Return([]*schema.Segmenter{&inactiveGlobalSegmenterOpenApi, &inactiveCustomSegmenterOpenApi}, nil)

	getSegmenterScopeGlobal := services.SegmenterScopeGlobal
	getSegmenterScopeProject := services.SegmenterScopeProject
	getSegmenterStatusActive := services.SegmenterStatusActive
	getSegmenterStatusInactive := services.SegmenterStatusInactive

	segmenterSvc.
		On(
			"ListSegmenters",
			int64(7),
			services.ListSegmentersParams{
				Scope:  &getSegmenterScopeGlobal,
				Status: &getSegmenterStatusActive,
			}).
		Return([]*schema.Segmenter{&activeGlobalSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(7),
			services.ListSegmentersParams{
				Scope:  &getSegmenterScopeProject,
				Status: &getSegmenterStatusActive,
			}).
		Return([]*schema.Segmenter{&activeCustomSegmenterOpenApi}, nil)
	segmenterSvc.
		On(
			"ListSegmenters",
			int64(7),
			services.ListSegmentersParams{
				Scope:  nil,
				Status: &getSegmenterStatusInactive,
			}).
		Return([]*schema.Segmenter{}, nil)
	segmenterSvc.
		On("ListSegmenters", int64(9), services.ListSegmentersParams{}).
		Return(nil, gorm.ErrRecordNotFound)
	segmenterSvc.
		On("GetSegmenter", int64(3), "nonexistent-segmenter").
		Return(nil, fmt.Errorf("unknown segmenter: nonexistent-segmenter"))
	segmenterSvc.
		On("GetSegmenter", int64(3), "test-global-segmenter").
		Return(&activeGlobalSegmenterOpenApi, nil)
	segmenterSvc.
		On("GetSegmenter", int64(4), "test-custom-segmenter").
		Return(&inactiveCustomSegmenterOpenApi, nil)
	segmenterSvc.
		On("GetSegmenter", int64(4), "test-new-custom-segmenter").
		Return(nil, fmt.Errorf("unknown segmenter: test-new-custom-segmenter"))
	segmenterSvc.
		On("GetCustomSegmenter", int64(4), "test-custom-segmenter").
		Return(&baseCustomSegmenter, nil)
	segmenterSvc.
		On("GetCustomSegmenter", int64(4), "test-missing-custom-segmenter").
		Return(nil, fmt.Errorf("unknown segmenter: test-missing-custom-segmenter"))
	segmenterSvc.
		On("GetCustomSegmenter", int64(5), "test-custom-segmenter").
		Return(&baseCustomSegmenter, nil)

	segmenterSvc.
		On(
			"CreateCustomSegmenter",
			int64(4),
			services.CreateCustomSegmenterRequestBody{
				Name: "already-existent-custom-segmenter",
			}).
		Return(
			nil,
			errors.Newf(
				errors.BadInput,
				"a segmenter with the name already-existent-custom-ensembler already exists",
			),
		)
	segmenterSvc.
		On(
			"CreateCustomSegmenter",
			int64(4),
			services.CreateCustomSegmenterRequestBody{
				Name: "test-new-custom-segmenter",
				Type: "STRING",
				Options: &models.Options{
					"option_a": true,
				},
				MultiValued: false,
				Constraints: &models.Constraints{
					{
						PreRequisites: []models.PreRequisite{
							{
								SegmenterName: "segmenter_name_1",
								SegmenterValues: []interface{}{
									"SN1",
								},
							},
						},
						AllowedValues: []interface{}{
							float64(1),
						},
						Options: &models.Options{
							"option_1": float64(1),
						},
					},
				},
				Required:    false,
				Description: &segmentersDescription,
			}).
		Return(&newCustomSegmenter, nil)

	segmenterSvc.
		On("UpdateCustomSegmenter",
			int64(4),
			"nonexistent-segmenter",
			services.UpdateCustomSegmenterRequestBody{}).
		Return(nil, fmt.Errorf("unknown segmenter: nonexistent-segmenter"))
	segmenterSvc.
		On("UpdateCustomSegmenter",
			int64(4),
			"test-custom-segmenter",
			services.UpdateCustomSegmenterRequestBody{
				Options: &models.Options{
					"option_a": true,
				},
				MultiValued: false,
				Constraints: &models.Constraints{
					{
						PreRequisites: []models.PreRequisite{
							{
								SegmenterName: "segmenter_name_1",
								SegmenterValues: []interface{}{
									"SN1",
								},
							},
						},
						AllowedValues: []interface{}{
							float64(1),
						},
						Options: &models.Options{
							"option_1": float64(1),
						},
					},
				},
				Required:    true,
				Description: &updatedDescription,
			}).
		Return(&updatedCustomSegmenter, nil)
	segmenterSvc.
		On("DeleteCustomSegmenter", int64(4), "nonexistent-segmenter").
		Return(fmt.Errorf("unknown segmenter: nonexistent-segmenter"))
	segmenterSvc.
		On("DeleteCustomSegmenter", int64(4), "test-custom-segmenter").
		Return(nil)

	settingsSvc := &mocks.ProjectSettingsService{}
	settingsSvc.
		On("GetDBRecord", models.ID(3)).
		Return(&projectSettings3, nil)
	settingsSvc.
		On("GetDBRecord", models.ID(4)).
		Return(&projectSettings4, nil)
	settingsSvc.
		On("GetDBRecord", models.ID(5)).
		Return(&projectSettings5, nil)
	settingsSvc.
		On("GetDBRecord", models.ID(6)).
		Return(&projectSettings6, nil)
	settingsSvc.
		On("GetDBRecord", models.ID(7)).
		Return(&projectSettings7, nil)
	settingsSvc.
		On("GetDBRecord", models.ID(8)).
		Return(nil, errors.Newf(errors.Unknown, "test find project settings error"))

	// return values are set as (nil, nil) for convenience since the non-error return values are not used explicitly
	settingsSvc.
		On("GetProjectSettings", int64(3)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(4)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(5)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(6)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(7)).
		Return(nil, nil)
	settingsSvc.
		On("GetProjectSettings", int64(8)).
		Return(nil, errors.Newf(errors.NotFound, "test get project settings error"))
	settingsSvc.
		On("GetProjectSettings", int64(9)).
		Return(nil, nil)

	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(3)).Return(nil, nil)
	mlpSvc.On(
		"GetProject", int64(2),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(2)))
	mlpSvc.On("GetProject", int64(4)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(5)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(6)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(7)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(8)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(9)).Return(nil, nil)

	// Create test controller
	s.ctrl = &SegmenterController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				MLPService:             mlpSvc,
				SegmenterService:       segmenterSvc,
				ProjectSettingsService: settingsSvc,
			},
		},
	}
}

func TestSegmenterController(t *testing.T) {
	suite.Run(t, new(SegmenterControllerTestSuite))
}

func (s *SegmenterControllerTestSuite) TestListSegmenters() {
	t := s.Suite.T()

	segmenterScopeGlobal := schema.SegmenterScopeGlobal
	segmenterScopeProject := schema.SegmenterScopeProject
	segmenterStatusActive := schema.SegmenterStatusActive
	segmenterStatusInactive := schema.SegmenterStatusInactive

	invalidScope := schema.SegmenterScope("invalid_scope")
	invalidStatus := schema.SegmenterStatus("invalid_status")

	tests := []struct {
		name      string
		projectID int64
		expected  string
		params    api.ListSegmentersParams
	}{
		{
			name:      "failure | mlp project not found",
			projectID: 2,
			expected:  s.expectedSegmentersResponses[12],
			params:    api.ListSegmentersParams{},
		},
		{
			name:      "failure | invalid scope param",
			projectID: 3,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 400,
				"\"scope passed is not a string representing segmenter scope\""),
			params: api.ListSegmentersParams{
				Scope: &invalidScope,
			},
		},
		{
			name:      "failure | invalid status param",
			projectID: 3,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 400,
				"\"status passed is not a string representing segmenter status\""),
			params: api.ListSegmentersParams{
				Status: &invalidStatus,
			},
		},
		{
			name:      "failure | error retrieving list of segmenters",
			projectID: 9,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 500,
				"\"record not found\""),
			params: api.ListSegmentersParams{},
		},
		{
			name:      "success | 1 active / 0 inactive global segmenter, 0 active / 0 inactive custom segmenter",
			projectID: 3,
			expected:  s.expectedSegmentersResponses[1],
			params:    api.ListSegmentersParams{},
		},
		{
			name:      "success | 0 active / 1 inactive global segmenter, 0 active / 0 inactive custom segmenter",
			projectID: 4,
			expected:  s.expectedSegmentersResponses[2],
			params:    api.ListSegmentersParams{},
		},
		{
			name:      "success | 0 active / 1 inactive global segmenter, 1 active / 0 inactive custom segmenter",
			projectID: 5,
			expected:  s.expectedSegmentersResponses[3],
			params:    api.ListSegmentersParams{},
		},
		{
			name:      "success | 0 active / 1 inactive global segmenter, 0 active / 1 inactive custom segmenter",
			projectID: 6,
			expected:  s.expectedSegmentersResponses[4],
			params:    api.ListSegmentersParams{},
		},
		{
			name: "success | 1 active / 0 inactive global segmenter, 1 active / 0 inactive custom segmenter " +
				"with search params global+active",
			projectID: 7,
			expected:  s.expectedSegmentersResponses[1],
			params: api.ListSegmentersParams{
				Scope:  &segmenterScopeGlobal,
				Status: &segmenterStatusActive,
			},
		},
		{
			name: "success | 1 active / 0 inactive global segmenter, 1 active / 0 inactive custom segmenter " +
				"with search params project+active",
			projectID: 7,
			expected:  s.expectedSegmentersResponses[5],
			params: api.ListSegmentersParams{
				Scope:  &segmenterScopeProject,
				Status: &segmenterStatusActive,
			},
		},
		{
			name: "success | 1 active / 0 inactive global segmenter, 1 active / 0 inactive custom segmenter " +
				"with search params inactive",
			projectID: 7,
			expected:  s.expectedSegmentersResponses[6],
			params: api.ListSegmentersParams{
				Status: &segmenterStatusInactive,
			},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListSegmenters(w, nil, data.projectID, data.params)
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

func (s *SegmenterControllerTestSuite) TestGetSegmenter() {
	t := s.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		segmenterName string
		expected      string
	}{
		{
			name:      "failure | mlp project not found",
			projectID: 2,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 404,
				"\"MLP Project info for id 2 not found in the cache\""),
		},
		{
			name:          "failure | no segmenter with matching name not found",
			projectID:     3,
			segmenterName: "nonexistent-segmenter",
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 500,
				"\"unknown segmenter: nonexistent-segmenter\""),
		},
		{
			name:          "success | global segmenter",
			projectID:     3,
			segmenterName: "test-global-segmenter",
			expected:      s.expectedSegmentersResponses[7],
		},
		{
			name:          "success | custom segmenter",
			projectID:     4,
			segmenterName: "test-custom-segmenter",
			expected:      s.expectedSegmentersResponses[8],
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetSegmenter(w, nil, data.projectID, data.segmenterName)
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

func (s *SegmenterControllerTestSuite) TestCreateSegmenter() {
	t := s.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		segmenterData string
		expected      string
	}{
		{
			name:          "failure | invalid json",
			projectID:     4,
			segmenterData: `{"name": "test-custom-segmenter}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 400,
				"\"unexpected EOF\", \"message\":\"unexpected EOF\""),
		},
		{
			name:          "failure | mlp project not found",
			projectID:     2,
			segmenterData: `{"name": "test-custom-segmenter"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 404,
				"\"MLP Project info for id 2 not found in the cache\""),
		},
		{
			name:          "failure | error creating custom segmenter",
			projectID:     4,
			segmenterData: `{"name": "already-existent-custom-segmenter"}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 400,
				"\"a segmenter with the name already-existent-custom-ensembler already exists\""),
		},
		{
			name:      "success | create custom segmenter",
			projectID: 4,
			segmenterData: `{
				"name": "test-new-custom-segmenter", 
				"type": "string",
				"options": {
					"option_a": true
				},
				"multi_valued": false,
				"description": "Test Desc",
				"constraints": [
					{
						"allowed_values": [
							1
						],
						"options": {
							"option_1": 1
						},
						"pre_requisites": [
							{
								"segmenter_name": "segmenter_name_1",
								"segmenter_values": [
									"SN1"
								]
							}
						]
					}
				],
				"required": false
			}`,
			expected: s.expectedSegmentersResponses[10],
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.segmenterData)))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.CreateSegmenter(w, req, data.projectID)
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

func (s *SegmenterControllerTestSuite) TestUpdateSegmenter() {
	t := s.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		segmenterName string
		segmenterData string
		expected      string
	}{
		{
			name:          "failure | invalid json",
			projectID:     4,
			segmenterName: "test-custom-segmenter",
			segmenterData: `{
				"type": "string,
				"options": null,
				"multi_valued": false,
				"constraints": null,
				"required": true
			}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 400,
				"\"invalid character '\\\\n' in string literal\""),
		},
		{
			name:          "failure | mlp project not found",
			projectID:     2,
			segmenterName: "test-custom-segmenter",
			segmenterData: `{
				"type": "string",
				"options": null,
				"multi_valued": false,
				"constraints": null,
				"required": true
			}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 404,
				"\"MLP Project info for id 2 not found in the cache\""),
		},
		{
			name:          "failure | error updating custom segmenter",
			projectID:     4,
			segmenterName: "nonexistent-segmenter",
			segmenterData: `{}`,
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 500,
				"\"unknown segmenter: nonexistent-segmenter\""),
		},
		{
			name:          "success | update custom segmenter",
			projectID:     4,
			segmenterName: "test-custom-segmenter",
			segmenterData: `{
				"type": "string",
				"options": {
					"option_a": true
				},
				"multi_valued": false,
				"constraints": [
					{
						"allowed_values": [
							1
						],
						"options": {
							"option_1": 1
						},
						"pre_requisites": [
							{
								"segmenter_name": "segmenter_name_1",
								"segmenter_values": [
									"SN1"
								]
							}
						]
					}
				],
				"required": true,
				"description": "updated"
			}`,
			expected: s.expectedSegmentersResponses[11],
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(data.segmenterData)))
			s.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.UpdateSegmenter(w, req, data.projectID, data.segmenterName)
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

func (s *SegmenterControllerTestSuite) TestDeleteSegmenter() {
	t := s.Suite.T()

	tests := []struct {
		name          string
		projectID     int64
		segmenterName string
		expected      string
	}{
		{
			name:          "failure | missing project settings",
			projectID:     8,
			segmenterName: "test-custom-segmenter",
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"Settings for project_id 8 cannot be retrieved: test get project settings error\""),
		},
		{
			name:          "failure | mlp project not found",
			projectID:     2,
			segmenterName: "test-missing-custom-segmenter",
			expected: fmt.Sprintf(s.expectedErrorResponseFormat,
				404, "\"MLP Project info for id 2 not found in the cache\""),
		},
		{
			name:          "failure | error deleting custom segmenter",
			projectID:     4,
			segmenterName: "nonexistent-segmenter",
			expected: fmt.Sprintf(s.expectedErrorResponseFormat, 500,
				"\"unknown segmenter: nonexistent-segmenter\""),
		},
		{
			name:          "success",
			projectID:     4,
			segmenterName: "test-custom-segmenter",
			expected:      s.expectedSegmentersResponses[9],
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.DeleteSegmenter(w, nil, data.projectID, data.segmenterName)
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
