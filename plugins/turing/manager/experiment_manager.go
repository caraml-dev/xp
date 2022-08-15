package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gojek/turing/engines/experiment/log"
	"github.com/gojek/turing/engines/experiment/manager"
	inproc "github.com/gojek/turing/engines/experiment/plugin/inproc/manager"
	"golang.org/x/oauth2/google"

	xpclient "github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/common/api/schema"
	_config "github.com/caraml-dev/xp/plugins/turing/config"
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

const xpExperimentConfigSchema = `[
  ["yup.object"], ["yup.required"],
  [
    "yup.shape",
    {
      "variables": [["yup.array"], ["yup.of", [["yup.object"], ["yup.shape",
        {
          "name": [["yup.string"], ["yup.required"]],
          "field": [["yup.string"], ["yup.required", "Field name is required"]],
          "field_source": [["yup.string"],
            ["yup.required"],
            ["yup.oneOf", ["header", "payload"], "One of the supported field sources should be selected"]]
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
	validate       *validator.Validate
	httpClient     *xpclient.ClientWithResponses
	RemoteUI       _config.RemoteUI       `validate:"required,dive"`
	RunnerDefaults _config.RunnerDefaults `validate:"required,dive"`
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

	// Create request parameter config
	params := []_config.RequestParameter{}
	for _, item := range config.Variables {
		params = append(params, _config.RequestParameter{
			Parameter: item.Name,
			Field:     item.Field,
			FieldSrc:  item.FieldSource,
		})
	}

	// Retrieve passkey using the API
	project, err := em.GetProject(config.ProjectID)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf(errorMsg, err.Error())
	}

	// Convert data to json
	bytes, err := json.Marshal(_config.ExperimentRunnerConfig{
		Endpoint:          em.RunnerDefaults.Endpoint,
		Timeout:           em.RunnerDefaults.Timeout,
		ProjectID:         config.ProjectID,
		Passkey:           project.Passkey,
		RequestParameters: params,
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
		validate:       validator.New(),
		httpClient:     client,
		RemoteUI:       config.RemoteUI,
		RunnerDefaults: config.RunnerDefaults,
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
