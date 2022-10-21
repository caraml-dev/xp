package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caraml-dev/turing/engines/experiment/log"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	inproc "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/manager"
	xpclient "github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/common/api/schema"
	_config "github.com/caraml-dev/xp/plugins/turing/config"
	treatmentconfig "github.com/caraml-dev/xp/treatment-service/config"
	"github.com/go-playground/validator/v10"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"golang.org/x/oauth2/google"
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

// Default scope for the Google Auth token used for the XP APIs
var googleOAuthScope = "https://www.googleapis.com/auth/userinfo.email"

// experimentManager implements manager.CustomExperimentManager interface
type experimentManager struct {
	validate                     *validator.Validate
	httpClient                   *xpclient.ClientWithResponses
	RemoteUI                     _config.RemoteUI                     `validate:"required,dive"`
	RunnerDefaults               _config.RunnerDefaults               `validate:"required,dive"`
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

	// Retrieve passkey using the API
	project, err := em.GetProject(config.ProjectID)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Retrieve treatment service configuration (shared with the management service) using the API
	treatmentServicePluginConfig, err := em.GetTreatmentServicePluginConfig()
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Store configs in the new treatment service config
	treatmentServiceConfig, err := em.MakeTreatmentServiceConfig(treatmentServicePluginConfig)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Convert data to json
	bytes, err := json.Marshal(_config.ExperimentRunnerConfig{
		Endpoint:               em.RunnerDefaults.Endpoint,
		Timeout:                em.RunnerDefaults.Timeout,
		ProjectID:              config.ProjectID,
		Passkey:                project.Passkey,
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

func (em *experimentManager) GetProject(projectID int) (*schema.ProjectSettings, error) {
	projectsErrorTpl := "Error retrieving project: %s"

	projectResponse, err := em.httpClient.GetProjectSettingsWithResponse(context.Background(), int64(projectID))
	if err != nil {
		return nil, err
	}

	// Handle possible errors
	if projectResponse.JSON404 != nil {
		return nil, fmt.Errorf(projectsErrorTpl, projectResponse.JSON404.Message)
	}
	if projectResponse.JSON500 != nil {
		return nil, fmt.Errorf(projectsErrorTpl, projectResponse.JSON500.Message)
	}
	if projectResponse.JSON200 == nil {
		return nil, fmt.Errorf(projectsErrorTpl, "empty response body")
	}

	return &projectResponse.JSON200.Data, nil
}

func (em *experimentManager) GetTreatmentServicePluginConfig() (*schema.TreatmentServicePluginConfig, error) {
	treatmentServicePluginConfigErrorTpl := "Error retrieving config: %s"

	treatmentServicePluginConfigResponse, err := em.httpClient.GetTreatmentServicePluginConfigWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	// Handle possible errors
	if treatmentServicePluginConfigResponse.JSON500 != nil {
		return nil, fmt.Errorf(treatmentServicePluginConfigErrorTpl, treatmentServicePluginConfigResponse.JSON500.Message)
	}
	if treatmentServicePluginConfigResponse.JSON200 == nil {
		return nil, fmt.Errorf(treatmentServicePluginConfigErrorTpl, "empty response body")
	}

	return &treatmentServicePluginConfigResponse.JSON200.Data, nil
}

func (em *experimentManager) MakeTreatmentServiceConfig(
	treatmentServicePluginConfig *schema.TreatmentServicePluginConfig,
) (*treatmentconfig.Config, error) {
	// Extract maxS2CellLevel and mixS2CellLevel from the segmenter configuration stored as a map[string]interface{}
	segmenterConfig := make(map[string]interface{})
	segmenterConfig["s2_ids"] = *treatmentServicePluginConfig.SegmenterConfig

	// Iterates through all Sentry config labels to cast them as the type interface{}
	sentryConfigLabels := make(map[string]string)
	for k, v := range *treatmentServicePluginConfig.SentryConfig.Labels {
		if castedV, ok := v.(string); ok {
			sentryConfigLabels[k] = castedV
		}
	}

	return &treatmentconfig.Config{
		Port:       em.TreatmentServicePluginConfig.Port,
		ProjectIds: em.TreatmentServicePluginConfig.ProjectIds,
		AssignedTreatmentLogger: treatmentconfig.AssignedTreatmentLoggerConfig{
			Kind:                 em.TreatmentServicePluginConfig.AssignedTreatmentLogger.Kind,
			QueueLength:          em.TreatmentServicePluginConfig.AssignedTreatmentLogger.QueueLength,
			FlushIntervalSeconds: em.TreatmentServicePluginConfig.AssignedTreatmentLogger.FlushIntervalSeconds,
			BQConfig: &treatmentconfig.BigqueryConfig{
				Project: em.TreatmentServicePluginConfig.AssignedTreatmentLogger.BQConfig.Project,
				Dataset: em.TreatmentServicePluginConfig.AssignedTreatmentLogger.BQConfig.Dataset,
				Table:   em.TreatmentServicePluginConfig.AssignedTreatmentLogger.BQConfig.Table,
			},
			KafkaConfig: &treatmentconfig.KafkaConfig{
				Brokers:          em.TreatmentServicePluginConfig.AssignedTreatmentLogger.KafkaConfig.Brokers,
				Topic:            em.TreatmentServicePluginConfig.AssignedTreatmentLogger.KafkaConfig.Topic,
				MaxMessageBytes:  em.TreatmentServicePluginConfig.AssignedTreatmentLogger.KafkaConfig.MaxMessageBytes,
				CompressionType:  em.TreatmentServicePluginConfig.AssignedTreatmentLogger.KafkaConfig.CompressionType,
				ConnectTimeoutMS: em.TreatmentServicePluginConfig.AssignedTreatmentLogger.KafkaConfig.ConnectTimeoutMS,
			},
		},
		DeploymentConfig: treatmentconfig.DeploymentConfig{
			EnvironmentType: em.TreatmentServicePluginConfig.DeploymentConfig.EnvironmentType,
			MaxGoRoutines:   em.TreatmentServicePluginConfig.DeploymentConfig.MaxGoRoutines,
		},
		ManagementService: treatmentconfig.ManagementServiceConfig{
			URL:                  em.TreatmentServicePluginConfig.ManagementService.URL,
			AuthorizationEnabled: em.TreatmentServicePluginConfig.ManagementService.AuthorizationEnabled,
		},
		MonitoringConfig: treatmentconfig.Monitoring{
			Kind:         em.TreatmentServicePluginConfig.MonitoringConfig.Kind,
			MetricLabels: em.TreatmentServicePluginConfig.MonitoringConfig.MetricLabels,
		},
		SwaggerConfig: treatmentconfig.SwaggerConfig{
			Enabled:          em.TreatmentServicePluginConfig.SwaggerConfig.Enabled,
			AllowedOrigins:   em.TreatmentServicePluginConfig.SwaggerConfig.AllowedOrigins,
			OpenAPISpecsPath: em.TreatmentServicePluginConfig.SwaggerConfig.OpenAPISpecsPath,
		},
		NewRelicConfig: newrelic.Config{
			Enabled: *treatmentServicePluginConfig.NewRelicConfig.Enabled,
			AppName: *treatmentServicePluginConfig.NewRelicConfig.AppName,
		},
		PubSub: treatmentconfig.PubSub{
			Project:   *treatmentServicePluginConfig.PubSub.Project,
			TopicName: *treatmentServicePluginConfig.PubSub.TopicName,
		},
		SegmenterConfig: segmenterConfig,
		SentryConfig: sentry.Config{
			Enabled: *treatmentServicePluginConfig.SentryConfig.Enabled,
			Labels:  sentryConfigLabels,
		},
	}, nil
}

func NewExperimentManager(configData json.RawMessage) (manager.CustomExperimentManager, error) {
	var config _config.ExperimentManagerConfig

	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to create XP experiment manager: %s", err)
	}

	// Create Google Client
	googleClient, err := google.DefaultClient(context.Background(), googleOAuthScope)
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
		RunnerDefaults:               config.RunnerDefaults,
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
