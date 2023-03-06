package server

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/common/testutils"
	"github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement/service"
)

const (
	TOPIC = "updates"
)

type ManagementServiceTestSuite struct {
	suite.Suite
	projectSettings   []schema.ProjectSettings
	requestParameters map[int64][]string
	store             *service.InMemoryStore
	client            *management.ClientWithResponses
	server            *httptest.Server
	ctx               context.Context
	subscription      *pubsub.Subscription
	pubsubClient      *pubsub.Client
	emulator          testcontainers.Container
}

func randomString() string {
	return strconv.Itoa(rand.Int())
}

func randomBool() bool {
	return rand.Float32() > 0.5
}

func newRandomProjectSettings() schema.ProjectSettings {
	projectId := int64(rand.Int())
	randomSegmenter := randomString()
	segmentersMap := schema.ProjectSegmenters{
		Names: []string{randomSegmenter},
		Variables: schema.ProjectSegmenters_Variables{
			AdditionalProperties: map[string][]string{
				randomSegmenter: {randomSegmenter},
			},
		},
	}
	return schema.ProjectSettings{
		EnableS2idClustering: randomBool(),
		ProjectId:            projectId,
		Username:             randomString(),
		RandomizationKey:     randomString(),
		Segmenters:           segmentersMap,
	}
}

func newRandomProjectSettingsOfSize(size int) []schema.ProjectSettings {
	projectSettings := make([]schema.ProjectSettings, size)
	for i := 0; i < size; i++ {
		projectSettings[i] = newRandomProjectSettings()
	}
	return projectSettings
}

func newRandomExperiment(projectId int64) schema.Experiment {
	description := randomString()
	endTime := time.Now().Add(time.Duration(24) * time.Hour)
	interval := int32(rand.Int())
	name := randomString()
	startTime := time.Now()
	daysOfWeek := []interface{}{int64(rand.Intn(7))}
	hoursOfDay := []interface{}{int64(rand.Intn(24))}
	s2Ids := []interface{}{int64(rand.Int())}
	traffic := int32(rand.Int())
	experimentType := newRandomExperimentType()
	return schema.Experiment{
		ProjectId:   &projectId,
		Description: &description,
		EndTime:     &endTime,
		Interval:    &interval,
		Name:        &name,
		Segment: &schema.ExperimentSegment{
			"days_of_week": daysOfWeek,
			"hours_of_day": hoursOfDay,
			"s2_ids":       s2Ids,
		},
		StartTime: &startTime,
		Treatments: &[]schema.ExperimentTreatment{
			{
				Configuration: map[string]interface{}{
					"config": randomBool(),
				},
				Name:    randomString(),
				Traffic: &traffic,
			},
		},
		Type: &experimentType,
	}
}

func newRandomExperimentType() schema.ExperimentType {
	if randomBool() {
		return schema.ExperimentTypeAB
	}
	return schema.ExperimentTypeSwitchback
}

func getLastPublishedUpdate(ctx context.Context, timeout time.Duration, subscription *pubsub.Subscription) (*_pubsub.MessagePublishState, error) {
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

func (suite *ManagementServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.projectSettings = newRandomProjectSettingsOfSize(3)

	suite.requestParameters = map[int64][]string{}
	for _, settings := range suite.projectSettings {
		requestParams := service.ConvertExperimentVariablesType(settings.Segmenters.Variables)
		suite.requestParameters[settings.ProjectId] = requestParams
	}

	pubsubTestProject := "test"
	emulator, pubsubClient, err := testutils.StartPubSubEmulator(suite.ctx, pubsubTestProject)
	if err != nil {
		suite.FailNow("failed to start pub sub emulator", err.Error())
	}
	suite.pubsubClient = pubsubClient
	suite.emulator = emulator
	topics := []string{TOPIC}
	err = testutils.CreatePubsubTopic(pubsubClient, suite.ctx, topics)
	if err != nil {
		suite.FailNow("failed to create topics", err.Error())
	}

	queue, err := service.NewPubSubMessageQueue(service.PubSubConfig{
		GCPProject: pubsubTestProject,
		TopicName:  TOPIC,
	})
	if err != nil {
		suite.FailNow("fail to instantiate message queue", err.Error())
		return
	}
	segmentersType := map[string]schema.SegmenterType{}
	store, err := service.NewInMemoryStore(make([]schema.Experiment, 0), suite.projectSettings, queue, segmentersType)
	if err != nil {
		suite.FailNow("fail to instantiate experiment store", err.Error())
	}
	suite.store = store
	suite.server = NewServer(suite.store)
	client, err := management.NewClientWithResponses(suite.server.URL)
	if err != nil {
		suite.FailNow("fail to instantiate client", err.Error())
	}
	suite.client = client
}

func (suite *ManagementServiceTestSuite) SetupTest() {
	suite.store.Experiments = make([]schema.Experiment, 0)
	topic := suite.pubsubClient.Topic(TOPIC)
	subscriptionId := "sub-" + uuid.NewString()
	subscription, err := suite.pubsubClient.CreateSubscription(suite.ctx, subscriptionId, pubsub.SubscriptionConfig{
		Topic:            topic,
		ExpirationPolicy: time.Hour * 24,
	})
	if err != nil {
		suite.FailNow("fail to instantiate client", err.Error())
	}
	suite.subscription = subscription
}

func (suite *ManagementServiceTestSuite) TearDownSuite() {
	suite.server.Close()
	_ = suite.emulator.Terminate(suite.ctx)
}

func (suite *ManagementServiceTestSuite) TestListProjects() {
	response, err := suite.client.ListProjectsWithResponse(suite.ctx)
	projects := []schema.Project{}
	for _, settings := range suite.projectSettings {
		projects = append(projects, schema.Project{
			Id:               settings.ProjectId,
			CreatedAt:        settings.CreatedAt,
			UpdatedAt:        settings.UpdatedAt,
			Segmenters:       settings.Segmenters.Names,
			RandomizationKey: settings.RandomizationKey,
			Username:         settings.Username,
		})
	}
	suite.Require().NoError(err)
	suite.Require().Equal(projects, response.JSON200.Data)
}

func (suite *ManagementServiceTestSuite) TestGetRequestParameter() {
	response, err := suite.client.GetProjectExperimentVariablesWithResponse(suite.ctx, suite.projectSettings[0].ProjectId)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.requestParameters[suite.projectSettings[0].ProjectId], response.JSON200.Data)
}

func (suite *ManagementServiceTestSuite) TestExperimentCreation() {
	projectId := suite.projectSettings[0].ProjectId
	expectedExperiment := newRandomExperiment(projectId)
	updater := randomString()
	body := management.CreateExperimentJSONRequestBody{
		Description: expectedExperiment.Description,
		EndTime:     *expectedExperiment.EndTime,
		Interval:    expectedExperiment.Interval,
		Name:        *expectedExperiment.Name,
		Segment:     *expectedExperiment.Segment,
		StartTime:   *expectedExperiment.StartTime,
		Treatments:  *expectedExperiment.Treatments,
		Type:        *expectedExperiment.Type,
		UpdatedBy:   &updater,
	}
	response, err := suite.client.CreateExperimentWithResponse(suite.ctx, 1, body)
	suite.Require().Equal(http.StatusOK, response.StatusCode())
	suite.Require().NoError(err)
	createdExperiment := response.JSON200.Data
	suite.Require().Equal(expectedExperiment.Name, createdExperiment.Name)
	suite.Require().Len(suite.store.Experiments, 1)
	suite.Require().Equal(suite.store.Experiments[0].Name, createdExperiment.Name)
	suite.Require().Equal(schema.ExperimentStatusActive, *createdExperiment.Status)
	suite.Require().Equal(expectedExperiment.Treatments, createdExperiment.Treatments)
	suite.Require().Equal(updater, *createdExperiment.UpdatedBy)

	publishedUpdate, err := getLastPublishedUpdate(suite.ctx, 1*time.Second, suite.subscription)
	suite.Require().NoError(err)
	suite.Require().NotNil(publishedUpdate.Update)
	suite.Require().Equal(*expectedExperiment.Name, publishedUpdate.GetExperimentCreated().GetExperiment().Name)
}

func (suite *ManagementServiceTestSuite) TestListExperiment() {
	projectId := suite.projectSettings[0].ProjectId
	suite.store.Experiments = []schema.Experiment{newRandomExperiment(projectId)}
	params := management.ListExperimentsParams{}
	response, err := suite.client.ListExperimentsWithResponse(suite.ctx, projectId, &params)
	suite.Require().NoError(err)
	suite.Require().Len(suite.store.Experiments, 1)
	suite.Require().Equal(suite.store.Experiments[0].Treatments, response.JSON200.Data[0].Treatments)
}

func (suite *ManagementServiceTestSuite) TestGetExperiment() {
	projectId := suite.projectSettings[0].ProjectId
	experiment := newRandomExperiment(projectId)
	id := int64(rand.Int())
	experiment.Id = &id
	suite.store.Experiments = []schema.Experiment{experiment}
	response, err := suite.client.GetExperimentWithResponse(suite.ctx, projectId, *experiment.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(experiment.Treatments, response.JSON200.Data.Treatments)
}

func (suite *ManagementServiceTestSuite) TestExperimentUpdate() {
	projectId := suite.projectSettings[0].ProjectId
	experiment := newRandomExperiment(projectId)
	id := int64(rand.Int())
	experiment.Id = &id
	suite.store.Experiments = []schema.Experiment{experiment}
	newDescription := "updated"
	params := management.UpdateExperimentJSONRequestBody{
		Description: &newDescription,
	}
	response, err := suite.client.UpdateExperimentWithResponse(suite.ctx, projectId, *experiment.Id, params)
	suite.Require().NoError(err)
	suite.Require().Equal(params.Description, response.JSON200.Data.Description)
}

func (suite *ManagementServiceTestSuite) TestUpdateProjectSettings() {
	settings := suite.projectSettings[2]
	enableS2IdClustering := true
	segmenterName := "newsegmenter"
	variables := schema.ProjectSegmenters_Variables{AdditionalProperties: map[string][]string{
		segmenterName: {segmenterName},
	}}
	projectSegmenters := schema.ProjectSegmenters{
		Names:     []string{segmenterName},
		Variables: variables,
	}
	response, err := suite.client.UpdateProjectSettingsWithResponse(suite.ctx, settings.ProjectId, management.UpdateProjectSettingsJSONRequestBody{
		EnableS2idClustering: &enableS2IdClustering,
		RandomizationKey:     "newsegmenter",
		Segmenters:           projectSegmenters,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, response.StatusCode())
	suite.Require().Equal("newsegmenter", suite.store.ProjectSettings[2].RandomizationKey)

	publishedUpdate, err := getLastPublishedUpdate(suite.ctx, 1*time.Second, suite.subscription)
	suite.Require().NoError(err)
	suite.Require().NotNil(publishedUpdate.GetProjectSettingsUpdated())
	suite.Require().Equal("newsegmenter", publishedUpdate.GetProjectSettingsUpdated().ProjectSettings.RandomizationKey)
}

func (suite *ManagementServiceTestSuite) TestEnableExperiment() {
	projectId := suite.projectSettings[0].ProjectId
	experiment := newRandomExperiment(projectId)
	id := int64(rand.Int())
	experiment.Id = &id
	status := schema.ExperimentStatusInactive
	experiment.Status = &status
	suite.store.Experiments = []schema.Experiment{experiment}
	response, err := suite.client.EnableExperimentWithResponse(suite.ctx, projectId, *experiment.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, response.StatusCode())

	publishedUpdate, err := getLastPublishedUpdate(suite.ctx, 1*time.Second, suite.subscription)
	suite.Require().NoError(err)
	suite.Require().NotNil(publishedUpdate.Update)
	suite.Require().Equal(*experiment.Id, publishedUpdate.GetExperimentUpdated().GetExperiment().Id)
	suite.Require().Equal(_pubsub.Experiment_Active, publishedUpdate.GetExperimentUpdated().GetExperiment().Status)
}

func (suite *ManagementServiceTestSuite) TestDisableExperiment() {
	projectId := suite.projectSettings[0].ProjectId
	experiment := newRandomExperiment(projectId)
	id := int64(rand.Int())
	experiment.Id = &id
	status := schema.ExperimentStatusActive
	experiment.Status = &status
	suite.store.Experiments = []schema.Experiment{experiment}
	response, err := suite.client.DisableExperimentWithResponse(suite.ctx, projectId, *experiment.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, response.StatusCode())

	publishedUpdate, err := getLastPublishedUpdate(suite.ctx, 1*time.Second, suite.subscription)
	suite.Require().NoError(err)
	suite.Require().NotNil(publishedUpdate.Update)
	suite.Require().Equal(*experiment.Id, publishedUpdate.GetExperimentUpdated().GetExperiment().Id)
	suite.Require().Equal(_pubsub.Experiment_Inactive, publishedUpdate.GetExperimentUpdated().GetExperiment().Status)
}

func TestManagementServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ManagementServiceTestSuite))
}
