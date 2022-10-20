package controller

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/mocks"
	"github.com/stretchr/testify/suite"
)

type ConfigurationControllerTestSuite struct {
	suite.Suite
	ctrl *ConfigurationController

	expectedConfigurationResponse string
}

func (s *ConfigurationControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ConfigurationControllerTestSuite")

	s.expectedConfigurationResponse = "FDSAFS"

	newRelicAppName := "xp"
	newRelicEnabled := true

	pubSubConfigProject := "dev"
	pubSubConfigTopicName := "xp-update"

	maxS2CellLevel := 15
	minS2CellLevel := 14

	sentryConfigEnabled := false
	sentryConfigLabels := make(map[string]interface{})

	treatmentServiceConfiguration := schema.TreatmentServiceConfig{
		NewRelicConfig: &schema.NewRelicConfig{
			AppName: &newRelicAppName,
			Enabled: &newRelicEnabled,
		},
		PubSub: &schema.PubSub{
			Project:   &pubSubConfigProject,
			TopicName: &pubSubConfigTopicName,
		},
		SegmenterConfig: &schema.SegmenterConfig{
			S2Ids: &schema.S2Ids{
				MaxS2CellLevel: &maxS2CellLevel,
				MinS2CellLevel: &minS2CellLevel,
			},
		},
		SentryConfig: &schema.SentryConfig{
			Enabled: &sentryConfigEnabled,
			Labels:  &sentryConfigLabels,
		},
	}

	configurationSvc := &mocks.ConfigurationService{}
	configurationSvc.
		On("GetTreatmentServiceConfig").
		Return(treatmentServiceConfiguration, nil)

	// Create test controller
	s.ctrl = &ConfigurationController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				ConfigurationService: configurationSvc,
			},
		},
	}
}

func TestConfigurationController(t *testing.T) {
	suite.Run(t, new(ConfigurationControllerTestSuite))
}

func (s *ConfigurationControllerTestSuite) TestGetTreatmentServiceConfig() {
	w := httptest.NewRecorder()
	s.ctrl.GetTreatmentServiceConfig(w, nil)
	resp := w.Result()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().JSONEq(`{"data":{
		"new_relic_config":{
			"app_name":"xp",
			"enabled":true
		},
		"pub_sub":{
			"project":"dev",
			"topic_name":"xp-update"
		},
		"segmenter_config":{
			"s2_ids":{
				"max_s2_cell_level":15,
				"min_s2_cell_level":14
			}	
		},
		"sentry_config":{
			"enabled":false,
			"labels":{}
		}
	}}`, string(body))
}
