//go:build integration

package messagequeue_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/xp/common/api/schema"
	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/common/segmenters"
	common_testutils "github.com/caraml-dev/xp/common/testutils"
	tu "github.com/caraml-dev/xp/management-service/internal/testutils"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/messagequeue"
	"github.com/caraml-dev/xp/management-service/services/mocks"
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
	messagequeue.MessageQueueService
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
	db, cleanup, err := tu.CreateTestDB("file://../../database/db-migrations")
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
	messageQueueConfig := common_mq_config.MessageQueueConfig{
		Kind: "pubsub",
		PubSubConfig: &common_mq_config.PubSubConfig{
			Project:   PUBSUB_PROJECT,
			TopicName: PUBSUB_TOPIC,
		},
	}
	messageQueueService, err := messagequeue.NewMessageQueueService(messageQueueConfig)
	if err != nil {
		s.FailNow("failed to initialize message queue service", err.Error())
	}
	subscriptions, err := common_testutils.CreateSubscriptions(pubSubClient, s.ctx, topics)
	if err != nil {
		s.FailNow("failed to prepare subscriptions", err.Error())
	}
	s.subscriptions = subscriptions

	// Init mock segmenter service
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
		On("GetSegmenterTypes", int64(1)).
		Return(
			map[string]schema.SegmenterType{
				"string_segmenter":  schema.SegmenterTypeString,
				"integer_segmenter": schema.SegmenterTypeInteger,
				"float_segmenter":   schema.SegmenterTypeReal,
				"bool_segmenter":    schema.SegmenterTypeBool,
			},
			nil,
		)
	segmenterSvc.
		On("GetSegmenterTypes", int64(2)).
		Return(
			map[string]schema.SegmenterType{
				"string_segmenter":  schema.SegmenterTypeString,
				"integer_segmenter": schema.SegmenterTypeInteger,
				"float_segmenter":   schema.SegmenterTypeReal,
				"bool_segmenter":    schema.SegmenterTypeBool,
			},
			nil,
		)
	segmenterSvc.
		On("ValidateExperimentSegment", int64(1), mock.Anything, mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateExperimentVariables", int64(2), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidatePrereqSegmenters", int64(2), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("ValidateRequiredSegmenters", int64(2), mock.Anything).
		Return(nil)
	segmenterSvc.
		On("GetSegmenterConfigurations", int64(2), mock.Anything).
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
		MessageQueueService:      messageQueueService,
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
	s.MessageQueueService = allServices.MessageQueueService

	// Create experiment test data
	err = db.Create(&models.Settings{
		Config: &models.ExperimentationConfig{

			Segmenters: models.ProjectSegmenters{
				Names: []string{"string_segmenter"},
				Variables: map[string][]string{
					"string_segmenter": {"exp-var-1"},
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
		Version:     1,
	}, *expResponse)

	// Check Published Create message
	publishedUpdate, err := getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(experimentId, publishedUpdate.GetExperimentCreated().GetExperiment().Id)

	// Disable Experiment
	err = svc.DisableExperiment(projectId, experimentId)
	s.Suite.Require().NoError(err)

	// Check Published Update message
	publishedUpdate, err = getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
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
	publishedUpdate, err := getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
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

	publishedUpdate, err = getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(segmenters.Names, publishedUpdate.GetProjectSettingsUpdated().GetProjectSettings().Segmenters.Names)
	s.Suite.Require().Equal(
		segmenters.Variables["seg-5"],
		publishedUpdate.GetProjectSettingsUpdated().GetProjectSettings().Segmenters.Variables["seg-5"].Value)
}

func (s *PubSubServiceTestSuite) TestProjectSettingPublish() {
	segmenterName := "test-new-custom-segmenter"
	segmenterDescription := "test description"
	customSegmenter := models.CustomSegmenter{
		Name:        segmenterName,
		ProjectID:   1,
		Type:        models.SegmenterValueTypeString,
		Description: &segmenterDescription,
		Options: &models.Options{
			"option_a": "option_a_val",
		},
		Required: true,
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
					"1",
				},
				Options: &models.Options{
					"option_1": "option_1_val",
				},
			},
		},
	}
	expectedConstraint := []*segmenters.Constraint{
		{
			PreRequisites: []*segmenters.PreRequisite{
				{
					SegmenterName: "segmenter_name_1",
					SegmenterValues: &segmenters.ListSegmenterValue{
						Values: []*segmenters.SegmenterValue{
							{Value: &segmenters.SegmenterValue_String_{String_: "SN1"}},
						},
					},
				},
			},
			AllowedValues: &segmenters.ListSegmenterValue{
				Values: []*segmenters.SegmenterValue{
					{Value: &segmenters.SegmenterValue_String_{String_: "1"}},
				},
			},
			Options: map[string]*segmenters.SegmenterValue{
				"option_1": {Value: &segmenters.SegmenterValue_String_{String_: "option_1_val"}},
			},
		},
	}
	expectedOptions := map[string]*segmenters.SegmenterValue{
		"option_a": {Value: &segmenters.SegmenterValue_String_{String_: "option_a_val"}},
	}
	segmenterConfig, err := customSegmenter.GetConfiguration()
	s.Suite.Require().NoError(err)
	err = s.MessageQueueService.PublishProjectSegmenterMessage("create", segmenterConfig, 1)
	s.Suite.Require().NoError(err)
	publishedUpdate, err := getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	fetchSegmenter := publishedUpdate.GetProjectSegmenterCreated().GetProjectSegmenter()
	s.Suite.Require().Equal(customSegmenter.Name, fetchSegmenter.Name)
	s.Suite.Require().Equal(string(customSegmenter.Type), fetchSegmenter.Type.String())
	s.Suite.Require().Equal(*customSegmenter.Description, fetchSegmenter.Description)
	s.Suite.Require().Equal(customSegmenter.Required, fetchSegmenter.Required)
	s.Suite.Require().Equal(expectedOptions, fetchSegmenter.Options)
	s.Suite.Require().Equal(expectedConstraint, fetchSegmenter.Constraints)

	customSegmenter.Required = !customSegmenter.Required
	segmenterConfig, err = customSegmenter.GetConfiguration()
	s.Suite.Require().NoError(err)
	err = s.MessageQueueService.PublishProjectSegmenterMessage("update", segmenterConfig, 1)
	s.Suite.Require().NoError(err)
	publishedUpdate, err = getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	fetchSegmenter = publishedUpdate.GetProjectSegmenterUpdated().GetProjectSegmenter()
	s.Suite.Require().False(fetchSegmenter.Required)

	err = s.MessageQueueService.PublishProjectSegmenterMessage("delete", segmenterConfig, 1)
	s.Suite.Require().NoError(err)
	publishedUpdate, err = getLastPublishedUpdate(s.ctx, 1*time.Second, s.subscriptions[PUBSUB_TOPIC])
	s.Suite.Require().NoError(err)
	s.Suite.Require().NotNil(publishedUpdate.Update)
	s.Suite.Require().Equal(int64(1), publishedUpdate.GetProjectSegmenterDeleted().ProjectId)
	s.Suite.Require().Equal(segmenterName, publishedUpdate.GetProjectSegmenterDeleted().SegmenterName)
}

func getLastPublishedUpdate(
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
