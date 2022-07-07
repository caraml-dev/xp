//go:build integration
// +build integration

package services_test

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/errors"
	tu "github.com/gojek/xp/management-service/internal/testutils"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type ProjectSettingsServiceTestSuite struct {
	suite.Suite
	services.ProjectSettingsService
	CleanUpFunc             func()
	ProjectSettings         []models.Settings
	SegmenterConfigurations []*_segmenters.SegmenterConfiguration
}

func (s *ProjectSettingsServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ProjectSettingsServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	s.SegmenterConfigurations = []*_segmenters.SegmenterConfiguration{
		{
			Name:        "seg-3",
			MultiValued: false,
			Options:     make(map[string]*_segmenters.SegmenterValue),
		},
	}

	// Init mock experiment service
	expSvc := &mocks.ExperimentService{}
	expSvc.
		On("ListAllExperiments",
			models.ID(2),
			mock.Anything,
		).
		Return(nil, gorm.ErrRecordNotFound)
	expSvc.
		On("ListAllExperiments",
			models.ID(1),
			mock.Anything,
		).
		Return([]*models.Experiment{}, nil)
	expSvc.
		On("ValidatePairwiseExperimentOrthogonality",
			int64(1),
			mock.Anything,
			[]string{"seg1"},
		).
		Return(nil)
	expSvc.
		On("ValidateProjectExperimentSegmentersExist",
			int64(1),
			mock.Anything,
			[]string{"seg1"},
		).
		Return(
			errors.Newf(
				errors.BadInput,
				"Error validating segmenters required for active experiment dummy experiment: experiment "+
					"test-experiment requires segmenter: inexistent_segmenter",
			),
		)

	// Init mock segmenter service
	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("GetSegmenterConfigurations", int64(3), []string{"seg5", "seg6"}).
		Return(nil, nil)
	segmenterSvc.
		On("GetSegmenterConfigurations", int64(2), []string{"seg1"}).
		Return(nil, nil)
	segmenterSvc.
		On("GetSegmenterConfigurations", int64(1), []string{"seg1"}).
		Return(nil, nil)
	segmenterSvc.
		On("ListGlobalSegmentersNames").
		Return([]string{"seg-3"})
	segmenterSvc.
		On("ValidateRequiredSegmenters", int64(3), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateRequiredSegmenters", int64(2), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateRequiredSegmenters", int64(1), []string{"seg1"}).
		Return(nil)
	segmenterSvc.
		On("ValidatePrereqSegmenters", int64(3), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidatePrereqSegmenters", int64(2), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidatePrereqSegmenters", int64(1), []string{"seg1"}).
		Return(nil)
	segmenterSvc.On("ValidateExperimentVariables", int64(3), mock.Anything).Return(nil)
	segmenterSvc.On("ValidateExperimentVariables", int64(2), mock.Anything).Return(nil)
	segmenterSvc.On("ValidateExperimentVariables", int64(1), mock.Anything).Return(nil)

	// Init mock validation service
	validationSvc := &mocks.ValidationService{}
	s2idClusterEnabled := true
	validationSvc.On(
		"Validate",
		services.CreateProjectSettingsRequestBody{
			Username: "client-3",
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
			TreatmentSchema: &models.TreatmentSchema{
				Rules: make([]models.Rule, 0),
			},
			RandomizationKey:     "rand-3",
			EnableS2idClustering: &s2idClusterEnabled,
		},
	).Return(nil)
	validationSvc.On(
		"Validate",
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
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
			ValidationUrl:        nil,
			RandomizationKey:     "rand-4",
			EnableS2idClustering: nil,
		},
	).Return(nil)
	validationSvc.On(
		"Validate",
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp-var-1", "exp-var-2"},
				}},
			ValidationUrl:    nil,
			RandomizationKey: "rand-2",
		},
	).Return(nil)

	// Init mock pubsub service
	pubSubSvc := &mocks.PubSubPublisherService{}
	pubSubSvc.On(
		"PublishProjectSettingsMessage",
		"create",
		mock.Anything,
	).Return(nil)
	pubSubSvc.On(
		"PublishProjectSettingsMessage",
		"update",
		mock.Anything,
	).Return(nil)

	allServices := &services.Services{
		ExperimentService:      expSvc,
		ValidationService:      validationSvc,
		PubSubPublisherService: pubSubSvc,
		SegmenterService:       segmenterSvc,
	}

	// Init user service
	s.ProjectSettingsService = services.NewProjectSettingsService(allServices, db)

	// Create test data
	s.ProjectSettings, err = createTestUsers(db)
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *ProjectSettingsServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up ProjectSettingsServiceTestSuite")
	s.CleanUpFunc()
}

func TestProjectSettingsService(t *testing.T) {
	suite.Run(t, new(ProjectSettingsServiceTestSuite))
}

func (s *ProjectSettingsServiceTestSuite) TestProjectSettingsServiceGetIntegration() {
	settingsResponse, err := s.ProjectSettingsService.GetProjectSettings(1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.ProjectSettings[0], *settingsResponse)
	settingsResponse, err = s.ProjectSettingsService.GetProjectSettings(2)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.ProjectSettings[1], *settingsResponse)
	settingsResponse, err = s.ProjectSettingsService.GetProjectSettings(4)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.ProjectSettings[2], *settingsResponse)
}

func (s *ProjectSettingsServiceTestSuite) TestGetExperimentVariablesIntegration() {
	// Get request parameters for valid global segmenters, seg2 is duplicate, GetExperimentVariables should not return duplicate
	params, err := s.ProjectSettingsService.GetExperimentVariables(1)
	s.Suite.Require().NoError(err)
	assert.ElementsMatch(s.Suite.T(), []string{"exp-var-1", "exp-var-2", "exp-var-3", "rand-1"}, *params)
	// Get request parameters for valid global segmenters
	params, err = s.ProjectSettingsService.GetExperimentVariables(2)
	s.Suite.Require().NoError(err)
	assert.ElementsMatch(s.Suite.T(), []string{"exp-var-3", "exp-var-4", "rand-2"}, *params)
}

func (s *ProjectSettingsServiceTestSuite) TestProjectSettingsServiceListCreateUpdateIntegration() {
	// List Projects
	expectedProjects := []models.Project{
		{
			CreatedAt:        s.ProjectSettings[0].CreatedAt,
			Id:               s.ProjectSettings[0].ProjectID.ToApiSchema(),
			RandomizationKey: s.ProjectSettings[0].Config.RandomizationKey,
			Segmenters:       s.ProjectSettings[0].Config.Segmenters.Names,
			UpdatedAt:        s.ProjectSettings[0].UpdatedAt,
			Username:         s.ProjectSettings[0].Username,
		},
		{
			CreatedAt:        s.ProjectSettings[1].CreatedAt,
			Id:               s.ProjectSettings[1].ProjectID.ToApiSchema(),
			RandomizationKey: s.ProjectSettings[1].Config.RandomizationKey,
			Segmenters:       s.ProjectSettings[1].Config.Segmenters.Names,
			UpdatedAt:        s.ProjectSettings[1].UpdatedAt,
			Username:         s.ProjectSettings[1].Username,
		},
		{
			CreatedAt:        s.ProjectSettings[2].CreatedAt,
			Id:               s.ProjectSettings[2].ProjectID.ToApiSchema(),
			RandomizationKey: s.ProjectSettings[2].Config.RandomizationKey,
			Segmenters:       s.ProjectSettings[2].Config.Segmenters.Names,
			UpdatedAt:        s.ProjectSettings[2].UpdatedAt,
			Username:         s.ProjectSettings[2].Username,
		},
	}
	projectsResponse, err := s.ProjectSettingsService.ListProjects()
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), expectedProjects, *projectsResponse)

	// Create Settings
	projectId := int64(3)
	s2idClusterEnabled := true
	settingsResponse, err := s.ProjectSettingsService.CreateProjectSettings(
		projectId,
		services.CreateProjectSettingsRequestBody{
			Username: "client-3",
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
			TreatmentSchema: &models.TreatmentSchema{
				Rules: make([]models.Rule, 0),
			},
			RandomizationKey:     "rand-3",
			EnableS2idClustering: &s2idClusterEnabled,
		})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), models.Settings{
		Model: models.Model{
			CreatedAt: settingsResponse.CreatedAt, // Copy the timestamp from the result
			UpdatedAt: settingsResponse.UpdatedAt, // Copy the timestamp from the result
		},
		ProjectID: models.ID(projectId),
		Username:  "client-3",
		Passkey:   settingsResponse.Passkey, // Copy the passkey from the result
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
			RandomizationKey:      "rand-3",
			S2IDClusteringEnabled: true,
		},
		TreatmentSchema: &models.TreatmentSchema{
			Rules: []models.Rule{},
		},
	}, *settingsResponse)
	s.Suite.Require().True(len(settingsResponse.Passkey) == 32)

	// Update Settings
	settingsResponse, err = s.ProjectSettingsService.UpdateProjectSettings(
		projectId,
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
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
			ValidationUrl:    nil,
			RandomizationKey: "rand-4",
		})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), models.Settings{
		Model: models.Model{
			CreatedAt: settingsResponse.CreatedAt, // Copy the timestamp from the result
			UpdatedAt: settingsResponse.UpdatedAt, // Copy the timestamp from the result
		},
		ProjectID: models.ID(projectId),
		Username:  "client-3",
		Passkey:   settingsResponse.Passkey, // Copy the passkey from the result
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg5", "seg6"},
				Variables: map[string][]string{
					"seg5": {"exp-var-5"},
					"seg6": {"exp-var-6"},
				}},
			RandomizationKey:      "rand-4",
			S2IDClusteringEnabled: true,
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
	}, *settingsResponse)
}

func (s *ProjectSettingsServiceTestSuite) TestProjectSettingsServiceUpdateExperimentsNotFound() {
	// Update Settings
	settingsResponse, err := s.ProjectSettingsService.UpdateProjectSettings(
		int64(2),
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp-var-1", "exp-var-2"},
				}},
			ValidationUrl:    nil,
			RandomizationKey: "rand-2",
		})
	s.Suite.Assert().EqualError(err, "record not found")
	s.Suite.Require().Nil(settingsResponse)
}

func (s *ProjectSettingsServiceTestSuite) TestProjectSettingsServiceUpdateInvalidRequiredSegmenterRemoval() {
	// Update Settings
	settingsResponse, err := s.ProjectSettingsService.UpdateProjectSettings(
		int64(1),
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp-var-1", "exp-var-2"},
				}},
			ValidationUrl:    nil,
			RandomizationKey: "rand-2",
		})
	s.Suite.Assert().EqualError(err, "Error validating segmenters required for active experiment dummy experiment: "+
		"experiment test-experiment requires segmenter: inexistent_segmenter")
	s.Suite.Require().Nil(settingsResponse)
}

func createTestUsers(db *gorm.DB) ([]models.Settings, error) {
	testValidationUrl := "https://test-validation-url.io"
	// Set up test settings records
	settingsRecords := []models.Settings{
		{
			Model: models.Model{
				CreatedAt: time.Date(2020, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2020, 1, 2, 3, 3, 3, 0, time.UTC),
			},
			Username: "client-1",
			Passkey:  "passkey-1",
			Config: &models.ExperimentationConfig{
				Segmenters: models.ProjectSegmenters{
					Names: []string{"seg1", "seg2"},
					Variables: map[string][]string{
						"seg1": {"exp-var-1", "exp-var-2"},
						"seg2": {"exp-var-2", "exp-var-3"},
					}},
				RandomizationKey: "rand-1",
			},
			ProjectID: models.ID(1),
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
			Username: "client-2",
			Passkey:  "passkey-2",
			Config: &models.ExperimentationConfig{
				Segmenters: models.ProjectSegmenters{
					Names: []string{"seg3", "seg4"},
					Variables: map[string][]string{
						"seg3": {"exp-var-3"},
						"seg4": {"exp-var-4"},
					}},
				RandomizationKey:      "rand-2",
				S2IDClusteringEnabled: true,
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
			ProjectID:     models.ID(2),
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
			Username: "client-4",
			Passkey:  "passkey-4",
			Config: &models.ExperimentationConfig{
				Segmenters: models.ProjectSegmenters{
					Names: []string{"seg7", "seg8"},
					Variables: map[string][]string{
						"seg7": {"exp-var-7"},
						"seg8": {"exp-var-8"},
					}},
				RandomizationKey:      "rand-4",
				S2IDClusteringEnabled: true,
			},
			TreatmentSchema: &models.TreatmentSchema{
				Rules: make([]models.Rule, 0),
			},
			ValidationUrl: &testValidationUrl,
			ProjectID:     models.ID(4),
		},
	}

	// Create settings records
	for _, settings := range settingsRecords {
		err := db.Create(&settings).Error
		if err != nil {
			return []models.Settings{}, err
		}
	}

	// Return expected user responses
	return settingsRecords, nil
}
