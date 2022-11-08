package services_test

import (
	"testing"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/config"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/stretchr/testify/suite"
)

type ConfigurationServiceTestSuite struct {
	suite.Suite
	services.ConfigurationService
}

func (s *ConfigurationServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ConfigurationServiceTestSuite")

	cfg := config.Config{
		NewRelicConfig: newrelic.Config{
			Enabled:           true,
			AppName:           "xp",
			License:           "amazing-license",
			IgnoreStatusCodes: []int{403, 404, 405},
			Labels:            map[string]interface{}{"env": "dev"},
		},
		PubSubConfig: &config.PubSubConfig{
			Project:   "dev",
			TopicName: "xp-update",
		},
		SegmenterConfig: map[string]interface{}{
			"s2_ids": map[string]interface{}{
				"mins2celllevel": 14,
				"maxs2celllevel": 15,
			},
		},
		SentryConfig: sentry.Config{Enabled: false, Labels: make(map[string]string)},
	}

	// Init configuration service
	s.ConfigurationService = services.NewConfigurationService(&cfg)
}

func TestConfigurationService(t *testing.T) {
	suite.Run(t, new(ConfigurationServiceTestSuite))
}

func (s *ConfigurationServiceTestSuite) TestGetTreatmentServicePluginConfig() {
	newRelicAppName := "xp"
	newRelicEnabled := true

	pubSubConfigProject := "dev"
	pubSubConfigTopicName := "xp-update"

	sentryConfigEnabled := false
	sentryConfigLabels := make(map[string]interface{})

	expectedConfiguration := schema.TreatmentServiceConfig{
		NewRelicConfig: &schema.NewRelicConfig{
			AppName: &newRelicAppName,
			Enabled: &newRelicEnabled,
		},
		PubSub: &schema.PubSub{
			Project:   &pubSubConfigProject,
			TopicName: &pubSubConfigTopicName,
		},
		SegmenterConfig: &schema.SegmenterConfig{
			"s2_ids": map[string]interface{}{
				"mins2celllevel": 14,
				"maxs2celllevel": 15,
			},
		},
		SentryConfig: &schema.SentryConfig{
			Enabled: &sentryConfigEnabled,
			Labels:  &sentryConfigLabels,
		},
	}
	actualConfiguration := s.ConfigurationService.GetTreatmentServiceConfig()
	s.Suite.Assert().Equal(expectedConfiguration, actualConfiguration)
}
