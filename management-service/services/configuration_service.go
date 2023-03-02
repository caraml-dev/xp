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

	return &configurationService{
		treatmentServiceConfig: schema.TreatmentServiceConfig{
			PubSub: &schema.PubSub{
				Project:   &cfg.MessageQueueConfig.PubSubConfig.Project,
				TopicName: &cfg.MessageQueueConfig.PubSubConfig.TopicName,
			},
			SegmenterConfig: &segmenterConfig,
		},
	}
}

func (svc configurationService) GetTreatmentServiceConfig() schema.TreatmentServiceConfig {
	return svc.treatmentServiceConfig
}
