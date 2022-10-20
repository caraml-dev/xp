package services

import (
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/config"
)

type ConfigurationService interface {
	GetTreatmentServicePluginConfig() schema.TreatmentServicePluginConfig
}

type configurationService struct {
	treatmentServicePluginConfig schema.TreatmentServicePluginConfig
}

func NewConfigurationService(cfg *config.Config) ConfigurationService {
	// Extract maxS2CellLevel and mixS2CellLevel from the segmenter configuration stored as a map[string]interface{}
	var maxS2CellLevel int
	var minS2CellLevel int
	if val, ok := cfg.SegmenterConfig["s2_ids"]; ok {
		segmenterConfig := val.(map[string]interface{})
		if val, ok := segmenterConfig["maxs2celllevel"]; ok {
			maxS2CellLevel = val.(int)
		}
		if val, ok := segmenterConfig["mins2celllevel"]; ok {
			minS2CellLevel = val.(int)
		}
	}

	// Iterates through all Sentry config labels to cast them as the type interface{}
	sentryConfigLabels := make(map[string]interface{})
	for k, v := range cfg.SentryConfig.Labels {
		sentryConfigLabels[k] = v
	}

	return &configurationService{
		treatmentServicePluginConfig: schema.TreatmentServicePluginConfig{
			NewRelicConfig: &schema.NewRelicConfig{
				AppName: &cfg.NewRelicConfig.AppName,
				Enabled: &cfg.NewRelicConfig.Enabled,
			},
			PubSub: &schema.PubSub{
				Project:   &cfg.PubSubConfig.Project,
				TopicName: &cfg.PubSubConfig.TopicName,
			},
			SegmenterConfig: &schema.SegmenterConfig{
				S2Ids: &schema.S2Ids{
					MaxS2CellLevel: &maxS2CellLevel,
					MinS2CellLevel: &minS2CellLevel,
				},
			},
			SentryConfig: &schema.SentryConfig{
				Enabled: &cfg.SentryConfig.Enabled,
				Labels:  &sentryConfigLabels,
			},
		},
	}
}

func (svc configurationService) GetTreatmentServicePluginConfig() schema.TreatmentServicePluginConfig {
	return svc.treatmentServicePluginConfig
}
