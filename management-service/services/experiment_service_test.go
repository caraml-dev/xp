// +build integration

package services_test

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/errors"
	tu "github.com/gojek/xp/management-service/internal/testutils"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type ExperimentServiceTestSuite struct {
	suite.Suite
	services.ExperimentService
	ExperimentHistoryService *mocks.ExperimentHistoryService
	CleanUpFunc              func()

	Settings    models.Settings
	Experiments []*models.Experiment
}

func (s *ExperimentServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ExperimentServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init mock services
	segmenterSvc := setupMockSegmenterService()
	validationSvc := setupMockValidationService()
	pubSubSvc := setupMockPubSubService()
	configuredTreatmentSvc := setupMockTreatmentService()

	// Init experiment history svc, mock calls will be set up during the test
	s.ExperimentHistoryService = &mocks.ExperimentHistoryService{}

	allServices := &services.Services{
		TreatmentService:         configuredTreatmentSvc,
		ValidationService:        validationSvc,
		ExperimentHistoryService: s.ExperimentHistoryService,
		SegmenterService:         segmenterSvc,
		PubSubPublisherService:   pubSubSvc,
	}

	// Init experiment service
	s.ExperimentService = services.NewExperimentService(allServices, db)

	// Create test data
	s.Settings, s.Experiments, err = createTestExperiments(db)

	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *ExperimentServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up ExperimentServiceTestSuite")
	s.CleanUpFunc()
}

func TestExperimentService(t *testing.T) {
	suite.Run(t, new(ExperimentServiceTestSuite))
}

func (s *ExperimentServiceTestSuite) TestExperimentServiceGetIntegration() {
	expResponse, err := s.ExperimentService.GetExperiment(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.Experiments[0], expResponse)
}

func (s *ExperimentServiceTestSuite) TestExperimentServiceListCreateUpdateIntegration() {
	// Test list experiments first, since the create/update of experiments
	// could affect the results
	testListExperiments(s)
	testCreateUpdateExperiment(s)
}

func testListExperiments(s *ExperimentServiceTestSuite) {
	t := s.Suite.T()
	svc := s.ExperimentService

	testStartTime := time.Date(2020, 2, 2, 5, 5, 6, 0, time.UTC)
	testEndTime := time.Date(2020, 2, 3, 3, 5, 6, 0, time.UTC)
	testName := "test-exp-1"
	testStatus := models.ExperimentStatusActive
	testExpType := models.ExperimentTypeAB
	testPage := int32(1)
	testPageSize := int32(2)

	integerSegmenter := []string{"1"}
	integer2Segmenter := []string{"1", "10"}
	floatSegmenter := []string{"1.0"}
	float2Segmenter := []string{"1.0", "3.0"}
	stringSegmenter := []string{"seg-1"}
	boolSegmenter := []string{"true"}

	// All experiments under a settings
	expResponsesList, pagingResponse, err := svc.ListExperiments(1, services.ListExperimentsParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 1, Total: 3}, pagingResponse)
	tu.AssertEqualValues(t, s.Experiments, expResponsesList)

	// No experiments filtered
	expResponsesList, pagingResponse, err = svc.ListExperiments(2, services.ListExperimentsParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 0, Total: 0}, pagingResponse)
	tu.AssertEqualValues(t, []*models.Experiment{}, expResponsesList)

	// Filter by a single parameter
	expResponsesList, pagingResponse, err = svc.ListExperiments(1,
		services.ListExperimentsParams{Status: &testStatus},
	)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 1, Total: 2}, pagingResponse)
	tu.AssertEqualValues(t, []*models.Experiment{s.Experiments[0], s.Experiments[2]}, expResponsesList)

	// Filter by all parameters
	expResponsesList, pagingResponse, err = svc.ListExperiments(1, services.ListExperimentsParams{
		Type:      &testExpType,
		Status:    &testStatus,
		StartTime: &testStartTime,
		EndTime:   &testEndTime,
		Name:      &testName,
		PaginationOptions: pagination.PaginationOptions{
			Page:     &testPage,
			PageSize: &testPageSize,
		},
		Segment: models.ExperimentSegment{
			"integer_segmenter":   integerSegmenter,
			"integer_2_segmenter": integer2Segmenter,
			"float_segmenter":     floatSegmenter,
			"string_segmenter":    stringSegmenter,
			"bool_segmenter":      boolSegmenter,
		},
	})
	s.Suite.Require().NoError(err)
	s.Suite.Assert().Equal(&pagination.Paging{
		Page:  1,
		Pages: 1,
		Total: 1,
	}, pagingResponse)
	tu.AssertEqualValues(t, []*models.Experiment{s.Experiments[0]}, expResponsesList)

	// Use the same start and end times
	testExactTimestamp := time.Date(2021, 2, 2, 3, 5, 7, 0, time.UTC)
	expResponsesList, pagingResponse, err = svc.ListExperiments(1,
		services.ListExperimentsParams{
			StartTime: &testExactTimestamp,
			EndTime:   &testExactTimestamp,
		},
	)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 1, Total: 1}, pagingResponse)
	tu.AssertEqualValues(t, []*models.Experiment{s.Experiments[2]}, expResponsesList)

	// Partial match of segmenter on multiple experiments
	expResponsesList, pagingResponse, err = svc.ListExperiments(1,
		services.ListExperimentsParams{
			Segment: models.ExperimentSegment{
				"float_segmenter": float2Segmenter,
			},
		},
	)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 1, Total: 2}, pagingResponse)
	tu.AssertEqualValues(t,
		[]*models.Experiment{s.Experiments[0], s.Experiments[1]},
		expResponsesList,
	)

	// Weak match of segmenters
	expResponsesList, pagingResponse, err = svc.ListExperiments(1,
		services.ListExperimentsParams{
			Segment: models.ExperimentSegment{
				"float_segmenter": floatSegmenter,
			},
			IncludeWeakMatch: true,
		},
	)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, &pagination.Paging{Page: 1, Pages: 1, Total: 2}, pagingResponse)
	tu.AssertEqualValues(t,
		[]*models.Experiment{s.Experiments[0], s.Experiments[2]},
		expResponsesList,
	)
	// Match name or description
	testDesc := "-1"
	expResponsesList, _, err = svc.ListExperiments(1,
		services.ListExperimentsParams{
			Search: &testDesc,
		},
	)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, []*models.Experiment{s.Experiments[0], s.Experiments[2]}, expResponsesList)
}

func testCreateUpdateExperiment(s *ExperimentServiceTestSuite) {
	t := s.Suite.T()
	svc := s.ExperimentService

	// Create Experiment
	projectId := int64(1)
	experimentId := int64(4)
	description := "Test description"
	interval := int32(60)
	traffic := int32(100)
	rawStringSegmenter := []interface{}{"seg-1", "seg-2"}
	stringSegmenter := []string{"seg-1", "seg-2"}
	name := "treatment"
	config := map[string]interface{}{
		"weight": 0.2,
		"meta": map[string]interface{}{
			"created-by": "test",
		},
	}
	reqTreatments := []models.ExperimentTreatment{
		{
			Name:          name,
			Configuration: config,
			Traffic:       &traffic,
		},
	}
	segmentRaw := models.ExperimentSegmentRaw{
		"string_segmenter": rawStringSegmenter,
	}
	segment := models.ExperimentSegment{
		"string_segmenter": stringSegmenter,
	}
	updatedBy := "integration-test"
	expResponse, err := svc.CreateExperiment(s.Settings, services.CreateExperimentRequestBody{
		Description: &description,
		EndTime:     time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
		Interval:    &interval,
		Name:        "test-experiment-create",
		Segment:     segmentRaw,
		StartTime:   time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
		Status:      models.ExperimentStatusActive,
		Treatments:  reqTreatments,
		Type:        models.ExperimentTypeSwitchback,
		Tier:        models.ExperimentTierDefault,
		UpdatedBy:   &updatedBy,
	})
	s.Suite.Require().NoError(err)
	respTreatments := []models.ExperimentTreatment{
		{
			Name:          name,
			Configuration: config,
			Traffic:       &traffic,
		},
	}
	tu.AssertEqualValues(t, models.Experiment{
		ID: models.ID(experimentId),
		Model: models.Model{
			CreatedAt: expResponse.CreatedAt,
			UpdatedAt: expResponse.UpdatedAt,
		},
		ProjectID:   models.ID(projectId),
		Description: &description,
		EndTime:     time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
		Interval:    &interval,
		Name:        "test-experiment-create",
		Segment:     segment,
		StartTime:   time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
		Status:      models.ExperimentStatusActive,
		Treatments:  respTreatments,
		Tier:        models.ExperimentTierDefault,
		Type:        models.ExperimentTypeSwitchback,
		UpdatedBy:   updatedBy,
	}, *expResponse)

	// Update Experiment
	s.ExperimentHistoryService.On("CreateExperimentHistory", expResponse).Return(nil, nil)
	newDescription := "New Test description, tier"
	expResponse, err = svc.UpdateExperiment(s.Settings, experimentId, services.UpdateExperimentRequestBody{
		Description: &newDescription,
		EndTime:     time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
		Interval:    &interval,
		Segment:     segmentRaw,
		StartTime:   time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
		Status:      models.ExperimentStatusActive,
		Treatments:  reqTreatments,
		Type:        models.ExperimentTypeSwitchback,
		Tier:        models.ExperimentTierOverride,
		UpdatedBy:   &updatedBy,
	})
	s.Suite.Require().NoError(err)
	s.Suite.Assert().Equal(models.ID(4), expResponse.ID)
	s.Suite.Assert().Equal(&newDescription, expResponse.Description)
	s.Suite.Assert().Equal(models.ExperimentStatusActive, expResponse.Status)

	// Disable Experiment
	s.ExperimentHistoryService.On("CreateExperimentHistory", expResponse).Return(nil, nil)
	err = svc.DisableExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)
	exp, err := svc.GetExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)
	s.Suite.Require().Equal(models.ExperimentStatusInactive, exp.Status)

	// Enable Experiment
	s.ExperimentHistoryService.On("CreateExperimentHistory", exp).Return(nil, nil)
	err = svc.EnableExperiment(s.Settings, experimentId)
	s.Suite.Require().NoError(err)
	exp, err = svc.GetExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)
	s.Suite.Require().Equal(models.ExperimentStatusActive, exp.Status)
}

func (s *ExperimentServiceTestSuite) TestRunCustomValidation() {
	tests := map[string]struct {
		experiment    models.Experiment
		settings      models.Settings
		context       services.ValidationContext
		operationType services.OperationType
		errString     string
	}{
		"failure | incorrect value assertion from template schema rule returns error": {
			experiment: models.Experiment{
				Treatments: []models.ExperimentTreatment{
					{
						Configuration: map[string]interface{}{
							"field1": "abc",
							"field2": "def",
							"field3": map[string]interface{}{
								"field4": 0.1,
							},
						},
					},
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
				ValidationUrl: &successValidationUrl,
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Go template rule test-rule returns false",
		},
		"failure | validation url returns an error": {
			experiment: models.Experiment{
				Treatments: []models.ExperimentTreatment{
					{
						Configuration: map[string]interface{}{
							"field1": "abc",
							"field2": "def",
							"field3": map[string]interface{}{
								"field4": 0.1,
							},
							"field5": 1,
						},
					},
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{},
				},
				ValidationUrl: &failureValidationUrl,
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
			errString:     "Error validating data with validation URL: 500 Internal Server Error",
		},
		"success": {
			experiment: models.Experiment{
				Treatments: []models.ExperimentTreatment{
					{
						Configuration: map[string]interface{}{
							"field1": "abc",
							"field2": "def",
							"field3": map[string]interface{}{
								"field4": 0.1,
							},
							"field5": 1,
						},
					},
				},
			},
			settings: models.Settings{
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "test-rule-1",
							Predicate: "{{- (eq .field1 \"abc\") -}}",
						},
					},
				},
				ValidationUrl: &successValidationUrl,
			},
			context:       services.ValidationContext{},
			operationType: services.OperationTypeCreate,
		},
	}
	for name, test := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ExperimentService.RunCustomValidation(
				test.experiment,
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

func createTestExperiments(db *gorm.DB) (models.Settings, []*models.Experiment, error) {
	// Create test project settings (with project_id=1)
	var settings models.Settings
	err := db.Create(&models.Settings{
		ProjectID: models.ID(1),
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names:     []string{"seg-1"},
				Variables: nil,
			},
		},
	}).Error
	if err != nil {
		return settings, []*models.Experiment{}, err
	}
	// Query the created settings
	query := db.Where("project_id = 1").First(&settings)
	if err := query.Error; err != nil {
		return settings, []*models.Experiment{}, err
	}

	// Test segmenters
	integerSegmenter := []string{"1", "2"}
	integer2Segmenter := []string{"1", "2", "3", "4", "5"}
	floatSegmenter := []string{"1.0", "2.0"}
	float2Segmenter := []string{"3.0", "4.0"}
	stringSegmenter := []string{"seg-1"}
	boolSegmenter := []string{"true"}

	description := "test-desc-1"

	// Define test experiments
	experiments := []models.Experiment{
		{
			ProjectID:  models.ID(1),
			Name:       "test-exp-1",
			Type:       models.ExperimentTypeAB,
			Tier:       models.ExperimentTierDefault,
			Treatments: nil,
			Segment: models.ExperimentSegment{
				"integer_segmenter":   integerSegmenter,
				"integer_2_segmenter": integer2Segmenter,
				"float_segmenter":     floatSegmenter,
				"string_segmenter":    stringSegmenter,
				"bool_segmenter":      boolSegmenter,
			},
			Status:    models.ExperimentStatusActive,
			StartTime: time.Date(2020, 2, 2, 4, 5, 6, 0, time.UTC),
			EndTime:   time.Date(2020, 2, 3, 4, 5, 6, 0, time.UTC),
			Model: models.Model{
				CreatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
			},
		},
		{
			ProjectID:  models.ID(1),
			Name:       "test-exp-2",
			Type:       models.ExperimentTypeSwitchback,
			Tier:       models.ExperimentTierDefault,
			Treatments: nil,
			Segment: models.ExperimentSegment{
				"string_segmenter": stringSegmenter,
				"float_segmenter":  float2Segmenter,
			},
			Status:    models.ExperimentStatusInactive,
			StartTime: time.Date(2020, 2, 2, 4, 5, 6, 0, time.UTC),
			EndTime:   time.Date(2020, 2, 3, 4, 5, 6, 0, time.UTC),
			Model: models.Model{
				CreatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
			},
		},
		{
			ProjectID:  models.ID(1),
			Name:       "test-exp-3",
			Type:       models.ExperimentTypeAB,
			Tier:       models.ExperimentTierOverride,
			Treatments: nil,
			Segment: models.ExperimentSegment{
				"string_segmenter": stringSegmenter,
			},
			Status:    models.ExperimentStatusActive,
			StartTime: time.Date(2021, 2, 2, 3, 5, 7, 0, time.UTC),
			EndTime:   time.Date(2021, 2, 2, 3, 5, 8, 0, time.UTC),
			Model: models.Model{
				CreatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.UTC),
			},
			Description: &description,
		},
	}

	// Create test experiments
	for _, exp := range experiments {
		err := db.Create(&exp).Error
		if err != nil {
			return settings, []*models.Experiment{}, err
		}
	}

	// Return expected experiment responses
	experimentRecords := []*models.Experiment{
		&experiments[0], &experiments[1], &experiments[2],
	}
	experimentRecords[0].ID = models.ID(1)
	experimentRecords[1].ID = models.ID(2)
	experimentRecords[2].ID = models.ID(3)

	return settings, experimentRecords, nil
}

func setupMockSegmenterService() services.SegmenterService {
	rawStringSegmenter := []interface{}{"seg-1"}
	rawString2Segmenter := []interface{}{"seg-1", "seg-2"}
	rawIntegerSegmenter := []interface{}{float64(1)}
	rawInteger2Segmenter := []interface{}{float64(1), float64(10)}
	rawFloatSegmenter := []interface{}{float64(1)}
	rawFloat2Segmenter := []interface{}{float64(1), float64((3))}
	rawBoolSegmenter := []interface{}{true}

	stringSegmenter := []string{"seg-1"}
	string2Segmenter := []string{"seg-1", "seg-2"}

	respIntegerSegmenter := []interface{}{"1"}
	respInteger2Segmenter := []interface{}{"1", "10"}
	respFloatSegmenter := []interface{}{"1.0"}
	respFloat2Segmenter := []interface{}{"1.0", "3.0"}
	respStringSegmenter := []interface{}{"\"seg-1\""}
	respBoolSegmenter := []interface{}{"true"}

	createExpSegmentRaw := models.ExperimentSegmentRaw{
		"string_segmenter": rawString2Segmenter,
	}
	createExpSegment := models.ExperimentSegment{
		"string_segmenter": string2Segmenter,
	}
	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.On("GetFormattedSegmenters", models.ExperimentSegmentRaw(nil)).
		Return(map[string]*[]interface{}{}, nil)
	segmenterSvc.On("GetFormattedSegmenters", models.ExperimentSegmentRaw{}).
		Return(map[string]*[]interface{}{}, nil)
	segmenterSvc.
		On("GetFormattedSegmenters", models.ExperimentSegmentRaw{
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_2_segmenter": rawInteger2Segmenter,
			"float_segmenter":     rawFloatSegmenter,
			"string_segmenter":    rawStringSegmenter,
			"bool_segmenter":      rawBoolSegmenter,
		}).
		Return(map[string]*[]interface{}{
			"integer_segmenter":   &respIntegerSegmenter,
			"integer_2_segmenter": &respInteger2Segmenter,
			"float_segmenter":     &respFloatSegmenter,
			"string_segmenter":    &respStringSegmenter,
			"bool_segmenter":      &respBoolSegmenter,
		}, nil)
	segmenterSvc.
		On("GetFormattedSegmenters", models.ExperimentSegmentRaw{
			"float_segmenter": rawFloatSegmenter,
		}).
		Return(map[string]*[]interface{}{
			"float_segmenter": &respFloatSegmenter,
		}, nil)
	segmenterSvc.
		On("GetFormattedSegmenters", models.ExperimentSegmentRaw{
			"float_segmenter": rawFloat2Segmenter,
		}).
		Return(map[string]*[]interface{}{
			"float_segmenter": &respFloat2Segmenter,
		}, nil)
	segmenterSvc.
		On("GetSegmenterTypes").
		Return(
			map[string]schema.SegmenterType{
				"integer_segmenter":   schema.SegmenterTypeInteger,
				"integer_2_segmenter": schema.SegmenterTypeInteger,
				"float_segmenter":     schema.SegmenterTypeReal,
				"string_segmenter":    schema.SegmenterTypeString,
				"bool_segmenter":      schema.SegmenterTypeBool,
			},
		)
	segmenterSvc.
		On("ValidateExperimentSegment", []string{
			"seg-1",
		}, createExpSegmentRaw).
		Return(nil)

	desc := "test-desc-1"
	segmenterSvc.
		On("ValidateSegmentOrthogonality", []string{
			"seg-1",
		}, createExpSegmentRaw,
			[]models.Experiment{
				{
					ID:          models.ID(3),
					ProjectID:   models.ID(1),
					Name:        "test-exp-3",
					Description: &desc,
					Type:        models.ExperimentTypeAB,
					Tier:        models.ExperimentTierOverride,
					Status:      models.ExperimentStatusActive,
					Model: models.Model{
						// time.FixedZone is used to bypass fields defined in Postgres as timestamp with NO timezone,
						// where offset 0 is treated as UTC
						// Relevant SO: https://github.com/lib/pq/issues/329
						CreatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.FixedZone("", 0)),
						UpdatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.FixedZone("", 0)),
					},
					Segment: models.ExperimentSegment{
						"string_segmenter": stringSegmenter,
					},
					Treatments: nil,
					StartTime:  time.Date(2021, 2, 2, 3, 5, 7, 0, time.UTC),
					EndTime:    time.Date(2021, 2, 2, 3, 5, 8, 0, time.UTC),
				},
			}).
		Return(nil)
	segmenterSvc.
		On("ValidateSegmentOrthogonality", []string{
			"seg-1",
		}, createExpSegment,
			[]models.Experiment{
				{
					ID:          models.ID(4),
					ProjectID:   models.ID(1),
					Name:        "test-experiment-create",
					Description: &desc,
					Type:        models.ExperimentTypeSwitchback,
					Tier:        models.ExperimentTierDefault,
					Status:      models.ExperimentStatusActive,
					Model: models.Model{
						// time.FixedZone is used to bypass fields defined in Postgres as timestamp with NO timezone,
						// where offset 0 is treated as UTC
						// Relevant SO: https://github.com/lib/pq/issues/329
						CreatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.FixedZone("", 0)),
						UpdatedAt: time.Date(2020, 4, 1, 4, 5, 6, 0, time.FixedZone("", 0)),
					},
					Segment: models.ExperimentSegment{
						"string_segmenter": stringSegmenter,
					},
					Treatments: nil,
					StartTime:  time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
					EndTime:    time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
				},
			}).
		Return(nil)
	return segmenterSvc
}

func setupMockValidationService() services.ValidationService {
	validationSvc := &mocks.ValidationService{}
	description := "Test description"
	interval := int32(60)
	traffic := int32(100)
	rawStringSegmenter := []interface{}{"seg-1", "seg-2"}
	name := "treatment"
	config := map[string]interface{}{
		"weight": 0.2,
		"meta": map[string]interface{}{
			"created-by": "test",
		},
	}
	treatments := []models.ExperimentTreatment{
		{
			Name:          name,
			Configuration: config,
			Traffic:       &traffic,
		},
	}
	segmentRaw := models.ExperimentSegmentRaw{
		"string_segmenter": rawStringSegmenter,
	}
	updatedBy := "integration-test"
	validationSvc.On(
		"Validate",
		services.CreateExperimentRequestBody{
			Description: &description,
			EndTime:     time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
			Interval:    &interval,
			Name:        "test-experiment-create",
			Segment:     segmentRaw,
			StartTime:   time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
			Status:      models.ExperimentStatusActive,
			Treatments:  treatments,
			Tier:        models.ExperimentTierDefault,
			Type:        models.ExperimentTypeSwitchback,
			UpdatedBy:   &updatedBy,
		},
	).Return(nil)

	newDescription := "New Test description, tier"
	validationSvc.On(
		"Validate",
		services.UpdateExperimentRequestBody{
			Description: &newDescription,
			EndTime:     time.Date(2021, 2, 2, 4, 0, 0, 0, time.UTC),
			Interval:    &interval,
			Segment:     segmentRaw,
			StartTime:   time.Date(2021, 2, 2, 3, 0, 0, 0, time.UTC),
			Status:      models.ExperimentStatusActive,
			Treatments:  treatments,
			Tier:        models.ExperimentTierOverride,
			Type:        models.ExperimentTypeSwitchback,
			UpdatedBy:   &updatedBy,
		},
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeExperiment,
		mock.Anything,
		services.ValidationContext{},
		&successValidationUrl,
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeExperiment,
		mock.Anything,
		services.ValidationContext{},
		&failureValidationUrl,
	).Return(errors.Newf(errors.BadInput, "Error validating data with validation URL: 500 Internal Server Error"))

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeCreate,
		services.EntityTypeExperiment,
		mock.Anything,
		services.ValidationContext{},
		(*string)(nil),
	).Return(nil)

	validationSvc.On(
		"ValidateEntityWithExternalUrl",
		services.OperationTypeUpdate,
		services.EntityTypeExperiment,
		mock.Anything,
		mock.Anything,
		(*string)(nil),
	).Return(nil)

	return validationSvc
}

func setupMockPubSubService() services.PubSubPublisherService {
	pubSubSvc := &mocks.PubSubPublisherService{}
	pubSubSvc.On(
		"PublishExperimentMessage",
		"create",
		mock.Anything,
	).Return(nil)
	pubSubSvc.On(
		"PublishExperimentMessage",
		"update",
		mock.Anything,
	).Return(nil)
	return pubSubSvc
}

func setupMockTreatmentService() services.TreatmentService {
	treatmentSvc := &mocks.TreatmentService{}
	treatmentSvc.On(
		"GetTreatmentNames",
		models.ID(1), []int64{1},
	).Return([]string{"std-treatment-1"}, nil)
	treatmentSvc.On(
		"GetTreatmentNames",
		models.ID(1), []int64{1, 3}, // Id 3 is not registered
	).Return([]string{"std-treatment-1"}, nil)

	treatmentSvc.On(
		"RunCustomValidation",
		models.Treatment{
			Name:      "treatment",
			ProjectID: 1,
			Configuration: map[string]interface{}{
				"weight": 0.2,
				"meta": map[string]interface{}{
					"created-by": "test",
				},
			},
		},
		mock.Anything,
		services.ValidationContext{},
		services.OperationTypeCreate,
	).Return(nil)

	treatmentSvc.On(
		"RunCustomValidation",
		models.Treatment{
			Name:      "treatment",
			ProjectID: 1,
			Configuration: map[string]interface{}{
				"weight": 0.2,
				"meta": map[string]interface{}{
					"created-by": "test",
				},
			},
		},
		mock.Anything,
		mock.Anything,
		services.OperationTypeUpdate,
	).Return(nil)

	return treatmentSvc
}
