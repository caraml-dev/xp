// +build integration

package services_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/protobuf/proto"

	"github.com/gojek/xp/common/api/schema"
	_pubsub "github.com/gojek/xp/common/pubsub"
	common_testutils "github.com/gojek/xp/common/testutils"
	"github.com/gojek/xp/management-service/config"
	tu "github.com/gojek/xp/management-service/internal/testutils"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

const (
	PUBSUB_PROJECT = "test"
	PUBSUB_TOPIC   = "update"
)

type PubSubServiceTestSuite struct {
	suite.Suite
	services.ExperimentService
	services.ProjectSettingsService
	services.TreatmentService
	services.PubSubPublisherService
	CleanUpFunc     func()
	Settings        models.Settings
	ProjectSettings []models.ProjectSettings
	ctx             context.Context
	emulator        testcontainers.Container
	subscriptions   map[string]*pubsub.Subscription
}

func (s *PubSubServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up PubSubServiceTestSuite")
	s.ctx = context.Background()

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Start Emulator and PubSub Client
	emulator, pubSubClient, err := common_testutils.StartPubSubEmulator(s.ctx, PUBSUB_PROJECT)
	if err != nil {
		s.FailNow("failed to start pub sub emulator", err.Error())
	}
	s.emulator = emulator
	topics := []string{PUBSUB_TOPIC}
	pubSubConfig := config.PubSubConfig{
		Project:   PUBSUB_PROJECT,
		TopicName: PUBSUB_TOPIC,
	}
	pubSubPublisher, err := services.NewPubSubPublisherService(&pubSubConfig)
	if err != nil {
		s.FailNow("failed to initialize pubsub publisher", err.Error())
	}
	subscriptions, err := common_testutils.CreateSubscriptions(pubSubClient, s.ctx, topics)
	if err != nil {
		s.FailNow("failed to prepare subscriptions", err.Error())
	}
	s.subscriptions = subscriptions

	// Init mock segmenter service
	stringSegmenter := []string{"seg-1"}
	string2Segmenter := []string{"seg-1", "seg-2"}
	createExpSegment := models.ExperimentSegment{
		"string_segmenter": string2Segmenter,
	}

	// Define Segmenters
	rawStringSegmenter := []interface{}{"seg-1"}
	rawString2Segmenter := []interface{}{"seg-1", "seg-2"}
	rawIntegerSegmenter := []interface{}{float64(1)}
	rawFloatSegmenter := []interface{}{float64(1)}
	rawFloat2Segmenter := []interface{}{float64(1), float64(3)}
	rawBoolSegmenter := []interface{}{true}

	respStringSegmenter := []interface{}{"seg-1"}
	respIntegerSegmenter := []interface{}{"1"}
	respFloatSegmenter := []interface{}{"1.0"}
	respFloat2Segmenter := []interface{}{"1.0", "3.0"}
	respBoolSegmenter := []interface{}{"true"}

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.On("GetFormattedSegmenters", models.ExperimentSegmentRaw{}).Return(map[string]*[]interface{}{}, nil)
	segmenterSvc.On("GetFormattedSegmenters", models.ExperimentSegmentRaw{
		"string_segmenter": rawString2Segmenter,
	}).Return(map[string]*[]interface{}{"string_segmenter": &rawString2Segmenter}, nil)
	segmenterSvc.On("GetFormattedSegmenters", models.ExperimentSegmentRaw(nil)).
		Return(map[string]*[]interface{}{}, nil)
	segmenterSvc.
		On("GetFormattedSegmenters", models.ExperimentSegmentRaw{
			"string_segmenter":  rawStringSegmenter,
			"integer_segmenter": rawIntegerSegmenter,
			"float_segmenter":   rawFloatSegmenter,
			"bool_segmenter":    rawBoolSegmenter,
		}).
		Return(map[string]*[]interface{}{
			"string_segmenter":  &respStringSegmenter,
			"integer_segmenter": &respIntegerSegmenter,
			"float_segmenter":   &respFloatSegmenter,
			"bool_segmenter":    &respBoolSegmenter,
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
				"string_segmenter":  schema.SegmenterTypeString,
				"integer_segmenter": schema.SegmenterTypeInteger,
				"float_segmenter":   schema.SegmenterTypeReal,
				"bool_segmenter":    schema.SegmenterTypeBool,
			},
		)
	segmenterSvc.
		On("ValidateExperimentSegment", mock.Anything, mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateExperimentVariables", mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateSegmentOrthogonality", []string{"seg-1"}, createExpSegment,
			[]models.ExperimentSegment{{"string_segmenter": stringSegmenter}}).
		Return(nil)
	segmenterSvc.
		On("ValidatePrereqSegmenters", mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateRequiredSegmenters", mock.Anything).
		Return(nil)
	segmenterSvc.
		On("GetSegmenterConfigurations", mock.Anything).
		Return(nil, nil)

	expHistSvc := &mocks.ExperimentHistoryService{}
	expHistSvc.On("CreateExperimentHistory", mock.Anything).Return(nil, nil)

	treatmentHistSvc := &mocks.TreatmentHistoryService{}
	treatmentHistSvc.On("CreateTreatmentHistory", mock.Anything).Return(nil, nil)

	// Init mock validation service
	validationSvc := &mocks.ValidationService{}
	validationSvc.On("Validate", mock.Anything).Return(nil)

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

	description := "Test description"
	interval := int32(60)
	traffic := int32(100)
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
	segment := models.ExperimentSegmentRaw{
		"string_segmenter": rawString2Segmenter,
	}
	updatedBy := "integration-test"
	validationSvc.On(
		"Validate",
		services.CreateExperimentRequestBody{
			Description: &description,
			EndTime:     time.Date(2021, 2, 2, 4, 5, 6, 0, time.UTC),
			Interval:    &interval,
			Name:        "test-experiment-create",
			Segment:     segment,
			StartTime:   time.Date(2021, 2, 2, 3, 5, 6, 0, time.UTC),
			Status:      models.ExperimentStatusActive,
			Treatments:  treatments,
			Type:        models.ExperimentTypeSwitchback,
			Tier:        models.ExperimentTierDefault,
			UpdatedBy:   &updatedBy,
		},
	).Return(nil)

	allServices := &services.Services{
		ExperimentService:        s.ExperimentService,
		ExperimentHistoryService: expHistSvc,
		SegmenterService:         segmenterSvc,
		ValidationService:        validationSvc,
		PubSubPublisherService:   pubSubPublisher,
		TreatmentHistoryService:  treatmentHistSvc,
	}

	// Init experiment service
	s.ExperimentService = services.NewExperimentService(allServices, db)

	// Init treatment service
	s.TreatmentService = services.NewTreatmentService(allServices, db)
	allServices.ExperimentService = s.ExperimentService
	allServices.TreatmentService = s.TreatmentService

	// Init project settings service
	s.ProjectSettingsService = services.NewProjectSettingsService(allServices, db)

	// Create experiment test data
	err = db.Create(&models.Settings{
		Config: &models.ExperimentationConfig{

			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp-var-1"},
				},
			},
		},
		ProjectID: models.ID(1),
	}).Error
	if err != nil {
		s.Suite.T().Fatalf("Could not create project settings: %v", err)
	}

	// Query the created project settings
	var settings models.Settings
	query := db.Where("project_id = 1").First(&settings)
	if err := query.Error; err != nil {
		s.Suite.T().Fatalf("Could not query created project settings: %v", err)
	}
	s.Settings = settings
}

func (s *PubSubServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up PubSubServiceTestSuite")
	_ = s.emulator.Terminate(s.ctx)
	s.CleanUpFunc()
}

func TestPubSubPublisherService(t *testing.T) {
	suite.Run(t, new(PubSubServiceTestSuite))
}

func (s *PubSubServiceTestSuite) TestExperimentServiceCreateUpdatePublish() {
	t := s.Suite.T()
	svc := s.ExperimentService

	// Create Experiment
	projectId := int64(1)
	experimentId := int64(1)
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
	segment := models.ExperimentSegment{
		"string_segmenter": stringSegmenter,
	}
	updatedBy := "integration-test"
	expResponse, err := svc.CreateExperiment(s.Settings, services.CreateExperimentRequestBody{
		Description: &description,
		EndTime:     time.Date(2021, 2, 2, 4, 5, 6, 0, time.UTC),
		Interval:    &interval,
		Name:        "test-experiment-create",
		Segment:     segmentRaw,
		StartTime:   time.Date(2021, 2, 2, 3, 5, 6, 0, time.UTC),
		Status:      models.ExperimentStatusActive,
		Treatments:  treatments,
		Type:        models.ExperimentTypeSwitchback,
		Tier:        models.ExperimentTierDefault,
		UpdatedBy:   &updatedBy,
	})
	s.Suite.Require().NoError(err)
	exp, err := svc.GetExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, models.Experiment{
		Model: models.Model{
			CreatedAt: exp.CreatedAt,
			UpdatedAt: exp.UpdatedAt,
		},
		ID:          models.ID(experimentId),
		ProjectID:   models.ID(projectId),
		Description: &description,
		EndTime:     time.Date(2021, 2, 2, 4, 5, 6, 0, time.UTC),
		Interval:    &interval,
		Name:        "test-experiment-create",
		Segment:     segment,
		StartTime:   time.Date(2021, 2, 2, 3, 5, 6, 0, time.UTC),
		Status:      models.ExperimentStatusActive,
		Treatments:  treatments,
		Type:        models.ExperimentTypeSwitchback,
		Tier:        models.ExperimentTierDefault,
		UpdatedBy:   "integration-test",
	}, *expResponse)

	// Check Published Create message
	publishedUpdate, err := getLastPublishedExperimentUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(experimentId, publishedUpdate.GetExperimentCreated().GetExperiment().Id)

	// Disable Experiment
	err = svc.DisableExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)

	// Check Published Update message
	publishedUpdate, err = getLastPublishedExperimentUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(_pubsub.Experiment_Inactive, publishedUpdate.GetExperimentUpdated().GetExperiment().GetStatus())
}

func (s *PubSubServiceTestSuite) TestProjectSettingsServiceCreateUpdatePublish() {
	// Create Project Settings
	projectId := int64(2)
	userName := "client-2"
	s2idClusterEnabled := true
	segmenters := models.ProjectSegmenters{
		Names: []string{"seg-5", "seg-6"},
		Variables: map[string][]string{
			"seg-5": {"exp-var-5.1", "exp-var-5.2"},
			"seg-6": {"exp-var-6"},
		},
	}
	randomizationKey := "rand-3"
	settingsResponse, err := s.ProjectSettingsService.CreateProjectSettings(
		projectId, services.CreateProjectSettingsRequestBody{
			Username: userName,
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg-5", "seg-6"},
				Variables: map[string][]string{
					"seg-5": {"exp-var-5.1", "exp-var-5.2"},
					"seg-6": {"exp-var-6"},
				},
			},
			RandomizationKey:     "rand-3",
			EnableS2idClustering: &s2idClusterEnabled,
		})
	s.Suite.Require().NoError(err)
	expectedUpdatedConfig := models.ExperimentationConfig{
		Segmenters:            segmenters,
		RandomizationKey:      randomizationKey,
		S2IDClusteringEnabled: s2idClusterEnabled,
	}
	s.Suite.Require().Equal(userName, settingsResponse.Username)
	s.Suite.Require().Equal(expectedUpdatedConfig, *settingsResponse.Config)

	// Check Published Create message
	publishedUpdate, err := getLastPublishedProjectSettingsUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(projectId, publishedUpdate.GetProjectSettingsCreated().GetProjectSettings().GetProjectId())

	// Update Project Settings
	newSegmenters := models.ProjectSegmenters{
		Names: []string{"seg-5"},
		Variables: map[string][]string{
			"seg-5": {"exp-var-5.1", "exp-var-5.2"},
		},
	}
	newRandomizationKey := "rand-4"
	settingsResponse, err = s.ProjectSettingsService.UpdateProjectSettings(projectId, services.UpdateProjectSettingsRequestBody{
		Segmenters:       newSegmenters,
		RandomizationKey: newRandomizationKey,
	})
	s.Suite.Require().NoError(err)
	segmenters = models.ProjectSegmenters{
		Names: []string{"seg-5"},
		Variables: map[string][]string{
			"seg-5": {"exp-var-5.1", "exp-var-5.2"},
		},
	}
	expectedUpdatedConfig = models.ExperimentationConfig{
		Segmenters:            newSegmenters,
		RandomizationKey:      newRandomizationKey,
		S2IDClusteringEnabled: s2idClusterEnabled,
	}
	s.Suite.Require().Equal(userName, settingsResponse.Username)
	s.Suite.Require().Equal(expectedUpdatedConfig, *settingsResponse.Config)

	publishedUpdate, err = getLastPublishedProjectSettingsUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(segmenters.Names, publishedUpdate.GetProjectSettingsUpdated().GetProjectSettings().Segmenters.Names)
	s.Suite.Require().Equal(
		segmenters.Variables["seg-5"],
		publishedUpdate.GetProjectSettingsUpdated().GetProjectSettings().Segmenters.Variables["seg-5"].Value)
}

func getLastPublishedExperimentUpdate(
	ctx context.Context,
	timeout time.Duration,
	subscription *pubsub.Subscription,
) (*_pubsub.MessagePublishState, error) {
	contextWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	update := _pubsub.MessagePublishState{}
	err := subscription.Receive(contextWithTimeout, func(ctx context.Context, msg *pubsub.Message) {
		_ = proto.Unmarshal(msg.Data, &update)
		msg.Ack()
	})
	if err != nil {
		return nil, err
	}
	return &update, nil
}

func getLastPublishedProjectSettingsUpdate(
	ctx context.Context,
	timeout time.Duration,
	subscription *pubsub.Subscription,
) (*_pubsub.MessagePublishState, error) {
	contextWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	update := _pubsub.MessagePublishState{}
	err := subscription.Receive(contextWithTimeout, func(ctx context.Context, msg *pubsub.Message) {
		_ = proto.Unmarshal(msg.Data, &update)
		msg.Ack()
	})
	if err != nil {
		return nil, err
	}
	return &update, nil
}
