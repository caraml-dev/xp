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
}

func (s *ConfigurationControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ConfigurationControllerTestSuite")

	pubSubConfigProject := "dev"
	pubSubConfigTopicName := "xp-update"

	treatmentServicePluginConfiguration := schema.TreatmentServiceConfig{
		PubSub: &schema.PubSub{
			Project:   &pubSubConfigProject,
			TopicName: &pubSubConfigTopicName,
		},
		SegmenterConfig: &schema.SegmenterConfig{
			"s2_ids": map[string]interface{}{
				"min_s2_cell_level": 14,
				"max_s2_cell_level": 15,
			},
		},
	}

	configurationSvc := &mocks.ConfigurationService{}
	configurationSvc.
		On("GetTreatmentServiceConfig").
		Return(treatmentServicePluginConfiguration, nil)

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

func (s *ConfigurationControllerTestSuite) TestGetTreatmentServicePluginConfig() {
	w := httptest.NewRecorder()
	s.ctrl.GetTreatmentServiceConfig(w, nil)
	resp := w.Result()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().JSONEq(`{"data":{
		"pub_sub":{
			"project":"dev",
			"topic_name":"xp-update"
		},
		"segmenter_config":{
			"s2_ids":{
				"max_s2_cell_level":15,
				"min_s2_cell_level":14
			}	
		}
	}}`, string(body))
}
