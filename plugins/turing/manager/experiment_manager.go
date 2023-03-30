package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/caraml-dev/turing/engines/experiment/log"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	inproc "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/manager"
	"github.com/gojek/mlp/api/pkg/auth"

	xpclient "github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/common/api/schema"
	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	_config "github.com/caraml-dev/xp/plugins/turing/config"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/go-playground/validator/v10"
)

func init() {
	err := inproc.Register("xp", func(config json.RawMessage) (manager.ExperimentManager, error) {
		return NewExperimentManager(config)
	})
	if err != nil {
		log.Panicf("failed to register XP experiment manager, %v", err)
	}
}

// Default timeout for requests made to the XP API server
const defaultRequestTimeout = time.Second * 5

// TODO: Move the validation within the microfrontend component instead of exposing the
// the yup validation config via the API. This also has other limitations - eg:
// can't conditionally evaluate if `field` has a value when `field_source` is not none.
const xpExperimentConfigSchema = `[
  ["yup.object"], ["yup.required"],
  [
    "yup.shape",
    {
      "variables": [["yup.array"], ["yup.of", [["yup.object"], ["yup.shape",
        {
          "name": [["yup.string"], ["yup.required"]],
          "field": ["yup.string"],
          "field_source": [["yup.string"],
            ["yup.required"],
            ["yup.oneOf", ["none", "header", "payload"], "One of the supported field sources should be selected"]]
        }
      ]]],
	  ["yup.required"]]
    }
  ]
]`

// experimentManager implements manager.CustomExperimentManager interface
type experimentManager struct {
	validate                     *validator.Validate
	httpClient                   *xpclient.ClientWithResponses
	RemoteUI                     _config.RemoteUI                     `validate:"required,dive"`
	TreatmentServicePluginConfig _config.TreatmentServicePluginConfig `validate:"required,dive"`
}

func (em *experimentManager) GetEngineInfo() (manager.Engine, error) {
	return manager.Engine{
		Name:        "xp",
		DisplayName: "Turing Experiments",
		Type:        manager.CustomExperimentManagerType,
		CustomExperimentManagerConfig: &manager.CustomExperimentManagerConfig{
			RemoteUI: manager.RemoteUI{
				Name:   em.RemoteUI.Name,
				URL:    em.RemoteUI.URL,
				Config: em.RemoteUI.Config,
			},
			ExperimentConfigSchema: xpExperimentConfigSchema,
		},
	}, nil
}

func (em *experimentManager) GetExperimentRunnerConfig(rawConfig json.RawMessage) (json.RawMessage, error) {
	errorMsg := "Error creating experiment runner config: %s"

	// Convert the raw config to the XP configuration type.
	config, err := getExperimentConfig(rawConfig)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Retrieve treatment service configuration (driven by the management service) using the API
	treatmentServicePluginConfig, err := em.GetTreatmentServiceConfigFromManagementService()
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Store configs in the new treatment service config
	treatmentServiceConfig, err := em.MakeTreatmentServicePluginConfig(treatmentServicePluginConfig, config.ProjectID)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Convert data to json
	bytes, err := json.Marshal(_config.ExperimentRunnerConfig{
		RequestParameters:      config.Variables,
		TreatmentServiceConfig: treatmentServiceConfig,
	})
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}
	return bytes, nil
}

func (em *experimentManager) ValidateExperimentConfig(rawConfig json.RawMessage) error {
	config, err := getExperimentConfig(rawConfig)
	if err != nil {
		return err
	}
	return em.validate.Struct(config)
}

func (em *experimentManager) GetTreatmentServiceConfigFromManagementService() (*schema.TreatmentServiceConfig, error) {
	treatmentServiceConfigErrorTpl := "Error retrieving config: %s"

	treatmentServiceConfigResponse, err := em.httpClient.GetTreatmentServiceConfigWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	// Handle possible errors
	if treatmentServiceConfigResponse.JSON500 != nil {
		return nil, fmt.Errorf(treatmentServiceConfigErrorTpl, treatmentServiceConfigResponse.JSON500.Message)
	}
	if treatmentServiceConfigResponse.JSON200 == nil {
		return nil, fmt.Errorf(treatmentServiceConfigErrorTpl, "empty response body")
	}

	return &treatmentServiceConfigResponse.JSON200.Data, nil
}

func (em *experimentManager) MakeTreatmentServicePluginConfig(
	treatmentServiceConfig *schema.TreatmentServiceConfig,
	projectID int,
) (*config.Config, error) {
	pluginConfig := &config.Config{
		Port:                    em.TreatmentServicePluginConfig.Port,
		ProjectIds:              []string{strconv.Itoa(projectID)},
		AssignedTreatmentLogger: em.TreatmentServicePluginConfig.AssignedTreatmentLogger,
		DebugConfig:             em.TreatmentServicePluginConfig.DebugConfig,
		DeploymentConfig:        em.TreatmentServicePluginConfig.DeploymentConfig,
		ManagementService:       em.TreatmentServicePluginConfig.ManagementService,
		MonitoringConfig:        em.TreatmentServicePluginConfig.MonitoringConfig,
		SwaggerConfig:           em.TreatmentServicePluginConfig.SwaggerConfig,
		NewRelicConfig:          em.TreatmentServicePluginConfig.NewRelicConfig,
		SentryConfig:            em.TreatmentServicePluginConfig.SentryConfig,
		SegmenterConfig:         *treatmentServiceConfig.SegmenterConfig,
	}
	messageQueueKind := *treatmentServiceConfig.MessageQueueConfig.Kind
	switch messageQueueKind {
	case schema.MessageQueueKindPubsub:
		pluginConfig.MessageQueueConfig = common_mq_config.MessageQueueConfig{
			Kind: "pubsub",
			PubSubConfig: &common_mq_config.PubSubConfig{
				Project:              *treatmentServiceConfig.MessageQueueConfig.PubSub.Project,
				TopicName:            *treatmentServiceConfig.MessageQueueConfig.PubSub.TopicName,
				PubSubTimeoutSeconds: em.TreatmentServicePluginConfig.PubSubTimeoutSeconds,
			},
		}
	case schema.MessageQueueKindNoop:
		pluginConfig.MessageQueueConfig = common_mq_config.MessageQueueConfig{
			Kind: "",
		}
	default:
		return nil, fmt.Errorf("invalid message queue kind (%s) was provided", messageQueueKind)
	}

	return pluginConfig, nil
}

func NewExperimentManager(configData json.RawMessage) (manager.CustomExperimentManager, error) {
	var config _config.ExperimentManagerConfig

	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to create XP experiment manager: %s", err)
	}

	// Create Google Client
	googleClient, err := auth.InitGoogleClient(context.Background())
	if err != nil {
		return nil, err
	}
	googleClient.Timeout = defaultRequestTimeout
	// Create XP client
	client, err := xpclient.NewClientWithResponses(
		config.BaseURL,
		xpclient.WithHTTPClient(googleClient),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to create XP management client: %s", err.Error())
	}

	em := &experimentManager{
		validate:                     _config.NewValidator(),
		httpClient:                   client,
		RemoteUI:                     config.RemoteUI,
		TreatmentServicePluginConfig: config.TreatmentServicePluginConfig,
	}

	err = em.validate.Struct(em)
	if err != nil {
		return nil, fmt.Errorf("failed to create XP experiment manager: %s", err)
	}

	return em, nil
}

func getExperimentConfig(rawConfig interface{}) (_config.ExperimentConfig, error) {
	// Using json marshal and unmarshal for flexibility in parsing the required values.
	var config _config.ExperimentConfig
	bytes, err := json.Marshal(rawConfig)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
