//go:build integration
// +build integration

package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/caraml-dev/xp/management-service/errors"
	tu "github.com/caraml-dev/xp/management-service/internal/testutils"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/mocks"
)

var successValidationUrl = "https://validation-success.io"

var failureValidationUrl = "https://validation-failure.io"

type TreatmentServiceTestSuite struct {
	suite.Suite
	services.TreatmentService
	TreatmentHistoryService *mocks.TreatmentHistoryService
	CleanUpFunc             func()

	Settings   models.Settings
	Treatments []*models.Treatment
}

func (s *TreatmentServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up TreatmentServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB(tu.MigrationsPath)
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init mock services
	validationSvc := setupMockTreatmentValidationService()

	// Init treatment history svc, mock calls will be set up during the test
	s.TreatmentHistoryService = &mocks.TreatmentHistoryService{}
	s.TreatmentHistoryService.On(
		"DeleteTreatmentHistory",
		int64(4),
	).Return(nil)

	allServices := &services.Services{
		ValidationService:       validationSvc,
		TreatmentHistoryService: s.TreatmentHistoryService,
	}

	// Init treatment service
	s.TreatmentService = services.NewTreatmentService(allServices, db)

	// Create test data
	s.Settings, s.Treatments, err = createTestTreatments(db)
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *TreatmentServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up TreatmentServiceTestSuite")
	s.CleanUpFunc()
}

func TestTreatmentService(t *testing.T) {
	suite.Run(t, new(TreatmentServiceTestSuite))
}

func (s *TreatmentServiceTestSuite) TestTreatmentServiceGetIntegration() {
	treatmentResponse, err := s.TreatmentService.GetTreatment(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.Treatments[0], treatmentResponse)
}

func (p *TreatmentServiceTestSuite) TestTreatmentServiceListCreateUpdateDeleteIntegration() {
	// Test list treatments first, since the create/update/delete of treatments
	// could affect the results
	testListTreatments(p)
	testCreateUpdateDeleteTreatment(p)
}

func testListTreatments(p *TreatmentServiceTestSuite) {
	t := p.Suite.T()
	svc := p.TreatmentService

	testName := "test-treatment-1"
	testPage := int32(1)
	testPageSize := int32(2)
	expectedPagingOpts := &pagination.Paging{Page: 1, Pages: 1, Total: 3}

	// All treatments under default list params
	treatmentResponsesList, pagingResponse, err := svc.ListTreatments(1, services.ListTreatmentsParams{})
	p.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, expectedPagingOpts, pagingResponse)
	tu.AssertEqualValues(t, p.Treatments, treatmentResponsesList)

	// No treatments filtered
	expectedPagingOpts = &pagination.Paging{Page: 1, Pages: 0, Total: 0}
	treatmentResponsesList, pagingResponse, err = svc.ListTreatments(2, services.ListTreatmentsParams{})
	p.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, expectedPagingOpts, pagingResponse)
	tu.AssertEqualValues(t, []*models.Treatment{}, treatmentResponsesList)

	// Filter by a single parameter
	expectedPagingOpts = &pagination.Paging{Page: 1, Pages: 1, Total: 1}
	treatmentResponsesList, pagingResponse, err = svc.ListTreatments(1,
		services.ListTreatmentsParams{Search: &testName},
	)
	p.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, expectedPagingOpts, pagingResponse)
	tu.AssertEqualValues(t, []*models.Treatment{p.Treatments[0]}, treatmentResponsesList)

	// Filter by all parameters
	treatmentResponsesList, pagingResponse, err = svc.ListTreatments(1, services.ListTreatmentsParams{
		Search: &testName,
		PaginationOptions: pagination.PaginationOptions{
			Page:     &testPage,
			PageSize: &testPageSize,
		},
	})
	p.Suite.Require().NoError(err)
	p.Suite.Assert().Equal(&pagination.Paging{
		Page:  1,
		Pages: 1,
		Total: 1,
	}, pagingResponse)
	tu.AssertEqualValues(t, []*models.Treatment{p.Treatments[0]}, treatmentResponsesList)
}

func testCreateUpdateDeleteTreatment(p *TreatmentServiceTestSuite) {
	t := p.Suite.T()
	svc := p.TreatmentService

	// Create Treatment
	updatedBy := "test-user"
	projectId := int64(1)
	treatmentId := int64(4)
	config := map[string]interface{}{"team": "business"}

	treatmentResponse, err := svc.CreateTreatment(p.Settings, services.CreateTreatmentRequestBody{
		Name:      "test-treatment-create",
		Config:    config,
		UpdatedBy: &updatedBy,
	})
	p.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, models.Treatment{
		ID: models.ID(treatmentId),
		Model: models.Model{
			CreatedAt: treatmentResponse.CreatedAt,
			UpdatedAt: treatmentResponse.UpdatedAt,
		},
		ProjectID:     models.ID(projectId),
		Name:          "test-treatment-create",
		Configuration: config,
		UpdatedBy:     updatedBy,
	}, *treatmentResponse)

	// Update Treatment
	p.TreatmentHistoryService.On("CreateTreatmentHistory", treatmentResponse).Return(nil, nil)
	newTreatmentConfig := models.TreatmentConfig{"team": "datascience"}
	treatmentResponse, err = svc.UpdateTreatment(p.Settings, treatmentId, services.UpdateTreatmentRequestBody{
		Config:    newTreatmentConfig,
		UpdatedBy: &updatedBy,
	})
	p.Suite.Require().NoError(err)
	p.Suite.Assert().Equal(models.ID(4), treatmentResponse.ID)
	p.Suite.Assert().Equal(newTreatmentConfig, treatmentResponse.Configuration)

	// Delete Treatment
	deletedTreatmentId := int64(4)
	err = svc.DeleteTreatment(projectId, deletedTreatmentId)
	p.Suite.Require().NoError(err)
}

func createTestTreatments(db *gorm.DB) (models.Settings, []*models.Treatment, error) {
	// Create test project settings (with project_id=1)
	var settings models.Settings
	err := db.Create(&models.Settings{
		ProjectID: models.ID(1),
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg-1"},
				Variables: map[string][]string{
					"seg-1": {"exp-var-1"},
				},
			},
		},
	}).Error
	if err != nil {
		return settings, []*models.Treatment{}, err
	}
	// Query the created settings
	query := db.Where("project_id = 1").First(&settings)
	if err := query.Error; err != nil {
		return settings, []*models.Treatment{}, err
	}

	// Define test treatments
	treatments := []models.Treatment{
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
			},
			ProjectID:     models.ID(1),
			Name:          "test-treatment-1",
			Configuration: map[string]interface{}{"team": "business"},
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
			},
			ProjectID:     models.ID(1),
			Name:          "test-treatment-2",
			Configuration: map[string]interface{}{"team": "business"},
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
			},
			ProjectID:     models.ID(1),
			Name:          "test-treatment-3",
			Configuration: map[string]interface{}{"team": "datascience"},
		},
	}

	// Create test treatments
	for _, treatment := range treatments {
		err := db.Create(&treatment).Error
		if err != nil {
			return settings, []*models.Treatment{}, err
		}
	}

	// Return expected treatment responses
	treatmentRecords := []*models.Treatment{
		&treatments[0], &treatments[1], &treatments[2],
	}
	treatmentRecords[0].ID = models.ID(1)
	treatmentRecords[1].ID = models.ID(2)
	treatmentRecords[2].ID = models.ID(3)

	return settings, treatmentRecords, nil
}

func setupMockTreatmentValidationService() services.ValidationService {
	validationSvc := &mocks.ValidationService{}
	updatedBy := "test-user"
	validationSvc.On(
		"Validate",
		services.CreateTreatmentRequestBody{
			Name:      "test-treatment-create",
			Config:    map[string]interface{}{"team": "business"},
			UpdatedBy: &updatedBy,
		},
	).Return(nil)
	validationSvc.On(
		"Validate",
		services.UpdateTreatmentRequestBody{
			Config:    map[string]interface{}{"team": "datascience"},
			UpdatedBy: &updatedBy,
		},
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeTreatment,
		map[string]interface{}{"team": "business"},
		services.ValidationContext{},
		(*string)(nil),
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeUpdate,
		services.EntityTypeTreatment,
		map[string]interface{}{"team": "datascience"},
		mock.Anything,
		(*string)(nil),
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeTreatment,
		map[string]interface{}{
			"field1": "abc",
			"field2": "def",
			"field3": map[string]interface{}{
				"field4": 0.1,
			},
		},
		services.ValidationContext{},
		(*string)(nil),
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeTreatment,
		map[string]interface{}{
			"field1": "abc",
			"field2": "def",
			"field3": map[string]interface{}{
				"field4": 0.1,
			},
			"field5": 1,
		},
		services.ValidationContext{},
		(*string)(nil),
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeTreatment,
		map[string]interface{}{
			"field": "failure-field",
		},
		services.ValidationContext{},
		&failureValidationUrl,
	).Return(errors.Newf(errors.BadInput, "Error validating data with validation URL: 500 Internal Server Error"))

	return validationSvc
}

func (s *TreatmentServiceTestSuite) TestRunCustomValidation() {
	tests := map[string]struct {
		treatmentConfig map[string]interface{}
		settings        models.Settings
		context         services.ValidationContext
		operationType   services.OperationType
		errString       string
	}{
		"failure | incorrect value assertion from template schema rule returns error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "test-rule",
							Predicate: "{{- (eq .field1 \"def\") -}}",
						},
					},
				},
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Go template rule test-rule returns false",
		},
		"failure | field with incompatible validation type in a treatment schema rule returns an error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
				"field5": 1,
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "test-rule",
							Predicate: "{{- (eq .field5 \"abc\") -}}",
						},
					},
				},
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString: "Error validating Go template rule test-rule: template: test-rule:1:5: executing \"test-rule\"" +
				" at <eq .field5 \"abc\">: error calling eq: incompatible types for comparison",
		},
		"failure | a template in the treatment schema returns neither true nor false": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "test-rule",
							Predicate: "{{- ( .field1) -}}",
						},
					},
				},
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Go template rule test-rule returns a value that is neither 'true' nor 'false': abc",
		},
		"failure | a rule in the treatment schema returns an error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "test-rule-1",
							Predicate: "{{- (eq .field1 \"abc\") -}}",
						},
						{
							Name:      "test-rule-2",
							Predicate: "{{- (eq .field2 \"def\") -}}",
						},
						{
							Name:      "test-rule-3",
							Predicate: "{{- (eq .field3.field4 0.2) -}}",
						},
					},
				},
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Go template rule test-rule-3 returns false",
		},
		"failure | validation url returns an error": {
			treatmentConfig: map[string]interface{}{
				"field": "failure-field",
			},
			settings: models.Settings{
				TreatmentSchema: nil,
				ValidationUrl:   &failureValidationUrl,
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Error validating data with validation URL: 500 Internal Server Error",
		},
		"success | no template schema specified": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			settings: models.Settings{
				TreatmentSchema: nil,
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
		},
		"success": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name: "test-rule",
							Predicate: "{{- or " +
								"(and\n " +
								"(eq .field1 \"abc\")\n " +
								"(eq .field2 \"def\")\n " +
								"(contains \"float\" (typeOf .field3.field4)))\n " +
								"(and\n " +
								"(eq .field1 \"xyz\")\n " +
								"(eq .field2 \"def\")\n " +
								"(contains \"int\" (typeOf .field3.field4))) " +
								"-}}",
						},
					},
				},
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
		},
	}
	for name, test := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.TreatmentService.RunCustomValidation(
				test.treatmentConfig,
				test.settings,
				test.context,
				test.operationType,
			)
			if test.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, test.errString)
			}
		})
	}
}

func (s *TreatmentServiceTestSuite) TestValidateTreatmentConfigWithTreatmentSchema() {
	tests := map[string]struct {
		treatmentConfig map[string]interface{}
		treatmentSchema *models.TreatmentSchema
		errString       string
	}{
		"failure | incorrect value assertion from template schema rule returns error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			treatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "test-rule",
						Predicate: "{{- (eq .field1 \"def\") -}}",
					},
				},
			},
			errString: "Go template rule test-rule returns false",
		},
		"failure | field with incompatible validation type in a treatment schema rule returns an error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
				"field5": 1,
			},
			treatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "test-rule",
						Predicate: "{{- (eq .field5 \"abc\") -}}",
					},
				},
			},
			errString: "Error validating Go template rule test-rule: template: test-rule:1:5: executing \"test-rule\"" +
				" at <eq .field5 \"abc\">: error calling eq: incompatible types for comparison",
		},
		"failure | a template in the treatment schema returns neither true nor false": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			treatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "test-rule",
						Predicate: "{{- ( .field1) -}}",
					},
				},
			},
			errString: "Go template rule test-rule returns a value that is neither 'true' nor 'false': abc",
		},
		"failure | a rule in the treatment schema returns an error": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			treatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name:      "test-rule-1",
						Predicate: "{{- (eq .field1 \"abc\") -}}",
					},
					{
						Name:      "test-rule-2",
						Predicate: "{{- (eq .field2 \"def\") -}}",
					},
					{
						Name:      "test-rule-3",
						Predicate: "{{- (eq .field3.field4 0.2) -}}",
					},
				},
			},
			errString: "Go template rule test-rule-3 returns false",
		},
		"success | no template schema specified": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			treatmentSchema: nil,
		},
		"success": {
			treatmentConfig: map[string]interface{}{
				"field1": "abc",
				"field2": "def",
				"field3": map[string]interface{}{
					"field4": 0.1,
				},
			},
			treatmentSchema: &models.TreatmentSchema{
				Rules: []models.Rule{
					{
						Name: "test-rule",
						Predicate: "{{- or " +
							"(and\n " +
							"(eq .field1 \"abc\")\n " +
							"(eq .field2 \"def\")\n " +
							"(contains \"float\" (typeOf .field3.field4)))\n " +
							"(and\n " +
							"(eq .field1 \"xyz\")\n " +
							"(eq .field2 \"def\")\n " +
							"(contains \"int\" (typeOf .field3.field4))) " +
							"-}}",
					},
				},
			},
		},
	}
	for name, test := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := services.ValidateTreatmentConfigWithTreatmentSchema(
				test.treatmentConfig,
				test.treatmentSchema,
			)
			if test.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, test.errString)
			}
		})
	}
}
