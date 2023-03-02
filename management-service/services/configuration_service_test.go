package services_test

import (
	"testing"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/config"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/stretchr/testify/suite"
)

type ConfigurationServiceTestSuite struct {
	suite.Suite
	services.ConfigurationService
}

func (s *ConfigurationServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ConfigurationServiceTestSuite")

	cfg := config.Config{
		MessageQueueConfig: &config.MessageQueueConfig{
			Kind: "pubsub",
			PubSubConfig: &config.PubSubConfig{
				Project:   "dev",
				TopicName: "xp-update",
			},
		},
		SegmenterConfig: map[string]interface{}{
			"s2_ids": map[string]interface{}{
				"mins2celllevel": 14,
				"maxs2celllevel": 15,
			},
		},
	}

	// Init configuration service
	s.ConfigurationService = services.NewConfigurationService(&cfg)
}

func TestConfigurationService(t *testing.T) {
	suite.Run(t, new(ConfigurationServiceTestSuite))
}

func (s *ConfigurationServiceTestSuite) TestGetTreatmentServicePluginConfig() {
	messageQueueKind := schema.MessageQueueKindPubsub
	pubSubConfigProject := "dev"
	pubSubConfigTopicName := "xp-update"

	expectedConfiguration := schema.TreatmentServiceConfig{
		MessageQueueConfig: &schema.MessageQueueConfig{
			Kind: &messageQueueKind,
			PubSub: &schema.PubSub{
				Project:   &pubSubConfigProject,
				TopicName: &pubSubConfigTopicName,
			},
		},
		SegmenterConfig: &schema.SegmenterConfig{
			"s2_ids": map[string]interface{}{
				"mins2celllevel": 14,
				"maxs2celllevel": 15,
			},
		},
	}
	actualConfiguration := s.ConfigurationService.GetTreatmentServiceConfig()
	s.Suite.Assert().Equal(expectedConfiguration, actualConfiguration)
}
