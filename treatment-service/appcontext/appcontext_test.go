package appcontext

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/common/testutils"
	"github.com/gojek/xp/treatment-service/config"
	"github.com/gojek/xp/treatment-service/models"
	"github.com/gojek/xp/treatment-service/monitoring"
	"github.com/gojek/xp/treatment-service/services"
	"github.com/gojek/xp/treatment-service/testhelper/mockmanagement/server"
	"github.com/gojek/xp/treatment-service/testhelper/mockmanagement/service"
)

func TestContext(t *testing.T) {

	// Config for emulator and test server
	pubSubConfig := config.PubSub{
		Project:              "test",
		TopicName:            "updates",
		PubSubTimeoutSeconds: 30,
	}
	projectSettings := []schema.ProjectSettings{
		{
			ProjectId: 1,
		},
	}
	segmentersType := map[string]schema.SegmenterType{
		"string_segmenter":  "string",
		"integer_segmenter": "integer",
		"float_segmenter":   "real",
		"bool_segmenter":    "bool",
	}

	// Setup emulator and test server
	emulator, testServer, err := SetupTest(pubSubConfig, projectSettings, segmentersType)
	if err != nil {
		assert.FailNow(t, "Test Setup fail", err.Error())
	}

	// Config setup for appcontext
	testConfig := config.Config{
		Port:                    0,
		ProjectIds:              []string{"1"},
		AssignedTreatmentLogger: config.AssignedTreatmentLoggerConfig{},
		NewRelicConfig:          newrelic.Config{},
		SentryConfig:            sentry.Config{},
		DeploymentConfig:        config.DeploymentConfig{},
		PubSub:                  pubSubConfig,
		ManagementService: config.ManagementServiceConfig{
			URL:                  testServer.URL,
			AuthorizationEnabled: false,
		},
		MonitoringConfig: config.Monitoring{},
		SwaggerConfig:    config.SwaggerConfig{},
		SegmenterConfig:  map[string]interface{}{"s2_ids": map[string]interface{}{"mins2celllevel": 10, "maxs2celllevel": 14}},
	}

	// Create appcontext
	appContext, err := NewAppContext(&testConfig)
	assert.NoError(t, err)

	// Create expected components less pubsub which cant be replicated due to context init
	localStorage, err := models.NewLocalStorage(
		testConfig.GetProjectIds(),
		testConfig.ManagementService.URL,
		testConfig.ManagementService.AuthorizationEnabled,
	)
	if err != nil {
		assert.FailNow(t, "error while creating local storage", err.Error())
	}
	segmenterSvc, err := services.NewSegmenterService(localStorage, testConfig.SegmenterConfig)
	if err != nil {
		assert.FailNow(t, "error while creating segmenter service", err.Error())
	}
	schemaSvc, err := services.NewSchemaService(localStorage, segmenterSvc)
	if err != nil {
		assert.FailNow(t, "error while creating schema service", err.Error())
	}
	assert.Equal(t, schemaSvc, appContext.SchemaService)
	experimentSvc, err := services.NewExperimentService(localStorage)
	if err != nil {
		assert.FailNow(t, "error while creating experiment service", err.Error())
	}
	assert.Equal(t, experimentSvc, appContext.ExperimentService)

	treatmentSvc, err := services.NewTreatmentService(localStorage)
	if err != nil {
		assert.FailNow(t, "error while creating treatment service", err.Error())
	}
	assert.Equal(t, treatmentSvc, appContext.TreatmentService)

	metricService, err := services.NewMetricService(testConfig.MonitoringConfig, localStorage)
	if err != nil {
		assert.FailNow(t, "error while creating metric service", err.Error())
	}
	assert.Equal(t, metricService, appContext.MetricService)

	logger, err := monitoring.NewNoopAssignedTreatmentLogger()
	if err != nil {
		assert.FailNow(t, "error while creating treatment logger", err.Error())
	}
	assert.Equal(t, logger, appContext.AssignedTreatmentLogger)

	TeardownTest(testServer, emulator)
}

func TeardownTest(testServer *httptest.Server, emulator testcontainers.Container) {
	ctx := context.Background()
	testServer.Close()
	err := emulator.Terminate(ctx)
	if err != nil {
		log.Fatal("Fail to shut down emulator gracefully")
	}
}

func SetupTest(
	pubSubConfig config.PubSub,
	projectSettings []schema.ProjectSettings,
	segmentersType map[string]schema.SegmenterType,
) (testcontainers.Container, *httptest.Server, error) {
	ctx := context.Background()
	emulator, pubsubClient, err := testutils.StartPubSubEmulator(ctx, pubSubConfig.Project)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start pub sub emulator")
	}
	topics := []string{pubSubConfig.TopicName}
	err = testutils.CreatePubsubTopic(pubsubClient, ctx, topics)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create topics")
	}
	queue, err := service.NewPubSubMessageQueue(service.PubSubConfig{
		GCPProject: pubSubConfig.Project,
		TopicName:  pubSubConfig.TopicName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("fail to instantiate message queue")
	}
	store, err := service.NewInMemoryStore(make([]schema.Experiment, 0), projectSettings, queue, segmentersType)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to instantiate experiment store")
	}
	testServer := server.NewServer(store)
	return emulator, testServer, err
}
