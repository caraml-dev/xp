package config

import (
	"fmt"
	"sync"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/go-playground/validator/v10"
)

type FieldSource string

const (
	// PayloadFieldSource is used to represent the request payload
	PayloadFieldSource FieldSource = "payload"
	// HeaderFieldSource is used to represent the request header
	HeaderFieldSource FieldSource = "header"
	// NoneFieldSource is used to represent that there is no field source,
	// i.e., the variable has not been configured.
	NoneFieldSource FieldSource = "none"
)

type RemoteUI struct {
	Name   string `json:"name" validate:"required"`
	URL    string `json:"url" validate:"required"`
	Config string `json:"config,omitempty"`
}

type ExperimentManagerConfig struct {
	Enabled                      bool                         `json:"enabled"`
	BaseURL                      string                       `json:"base_url"`      // Base URL for XP experiment REST API
	HomePageURL                  string                       `json:"home_page_url"` // Website URL for end-users to manage experiments
	RemoteUI                     RemoteUI                     `json:"remote_ui"`
	TreatmentServicePluginConfig TreatmentServicePluginConfig `json:"treatment_service_plugin_config"`
}

// ExperimentRunnerConfig is used to parse the XP runner config during initialization
type ExperimentRunnerConfig struct {
	RequestParameters      []Variable     `json:"request_parameters" validate:"required,dive"`
	TreatmentServiceConfig *config.Config `json:"treatment_service_config" validate:"required,dive"`
}

type TreatmentServicePluginConfig struct {
	Port                 int `json:"port" default:"8080"`
	PubSubTimeoutSeconds int `json:"pub_sub_timeout_seconds" validate:"required"`

	AssignedTreatmentLogger       config.AssignedTreatmentLoggerConfig `json:"assigned_treatment_logger"`
	DebugConfig                   config.DebugConfig                   `json:"debug_config"`
	DeploymentConfig              config.DeploymentConfig              `json:"deployment_config"`
	MessageQueueConfig            common_mq_config.MessageQueueConfig  `json:"message_queue_config"`
	ManagementService             config.ManagementServiceConfig       `json:"management_service"`
	MonitoringConfig              config.Monitoring                    `json:"monitoring_config"`
	SwaggerConfig                 config.SwaggerConfig                 `json:"swagger_config"`
	NewRelicConfig                newrelic.Config                      `json:"new_relic_config"`
	SentryConfig                  sentry.Config                        `json:"sentry_config"`
	ManagementServicePollerConfig config.ManagementServicePollerConfig `json:"management_service_poller_config"`
}

type Variable struct {
	Name        string      `json:"name" validate:"required"`
	Field       string      `json:"field"`
	FieldSource FieldSource `json:"field_source" validate:"required,oneof=none header payload"`
}

// ExperimentConfig is the experiment config saved on the Turing DB
type ExperimentConfig struct {
	// Keeping the ProjectID as a separate parameter for now, in the chance that
	// we need flexibility in the router -> experiment project association.
	ProjectID int        `json:"project_id"  validate:"required"`
	Variables []Variable `json:"variables" validate:"dive"`
}

// Custom validation for the Variable struct
func validateVariable(sl validator.StructLevel) {
	field := sl.Current().Interface().(Variable)
	if field.FieldSource == NoneFieldSource && field.Field != "" {
		sl.ReportError(field.Field, "Field", "name", "Value must not be set if FieldSource is none", "")
	} else if (field.FieldSource == HeaderFieldSource || field.FieldSource == PayloadFieldSource) && field.Field == "" {
		sl.ReportError(field.Field, "Field", "name", "Value must be set if FieldSource is not none", fmt.Sprintf("%v", field.Field))
	}
}

// Validate validates the fields in the ExperimentRunnerConfig for expected values
// and returns any errors
func (cfg *ExperimentRunnerConfig) Validate() error {
	validate := newExperimentRunnerConfigValidator()
	return validate.Struct(cfg)
}

// experimentRunnerConfigValidator is used to validate a given runner config
var experimentRunnerConfigValidator *validator.Validate
var expTreatmentValidatorOnce sync.Once

func newExperimentRunnerConfigValidator() *validator.Validate {
	expTreatmentValidatorOnce.Do(func() {
		// Save the validator to the global state
		experimentRunnerConfigValidator = NewValidator()
	})
	return experimentRunnerConfigValidator
}

func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterStructValidation(validateVariable, Variable{})
	return v
}
