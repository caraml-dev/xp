package services

import (
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/config"
)

type ConfigurationService interface {
	GetTreatmentServiceConfig() schema.TreatmentServiceConfig
}

type configurationService struct {
	treatmentServiceConfig schema.TreatmentServiceConfig
}

func NewConfigurationService(cfg *config.Config) ConfigurationService {
	var segmenterConfig schema.SegmenterConfig = cfg.SegmenterConfig

	var messageQueueKind schema.MessageQueueKind
	switch cfg.MessageQueueConfig.Kind {
	case "pubsub":
		messageQueueKind = schema.MessageQueueKindPubsub
	case "":
		messageQueueKind = schema.MessageQueueKindNoop
	}

	configurationSvc := &configurationService{
		treatmentServiceConfig: schema.TreatmentServiceConfig{
			MessageQueueConfig: &schema.MessageQueueConfig{
				Kind: &messageQueueKind,
			},
			SegmenterConfig: &segmenterConfig,
		},
	}
	if cfg.MessageQueueConfig.Kind == "pubsub" {
		configurationSvc.treatmentServiceConfig.MessageQueueConfig.PubSub = &schema.PubSub{
			Project:   &cfg.MessageQueueConfig.PubSubConfig.Project,
			TopicName: &cfg.MessageQueueConfig.PubSubConfig.TopicName,
		}
	}

	return configurationSvc
}

func (svc configurationService) GetTreatmentServiceConfig() schema.TreatmentServiceConfig {
	return svc.treatmentServiceConfig
}
