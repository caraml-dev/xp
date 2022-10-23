package config

import (
	"fmt"
	"sync"

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

type RunnerDefaults struct {
	Endpoint string `json:"endpoint"` // API host for the experiment runner
	Timeout  string `json:"timeout"`  // API timeout for the experiment runner
}

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
	RunnerDefaults               RunnerDefaults               `json:"runner_defaults"`
	TreatmentServicePluginConfig TreatmentServicePluginConfig `json:"treatment_service_plugin_config"`
}

// ExperimentRunnerConfig is used to parse the XP runner config during initialization
type ExperimentRunnerConfig struct {
	Endpoint               string                  `json:"endpoint" validate:"required"`
	ProjectID              int                     `json:"project_id" validate:"required"`
	Passkey                string                  `json:"passkey" validate:"required"`
	Timeout                string                  `json:"timeout" validate:"required"`
	RequestParameters      []Variable              `json:"request_parameters" validate:"required,dive"`
	TreatmentServiceConfig *TreatmentServiceConfig `json:"treatment_service_config"`
}

type TreatmentServicePluginConfig struct {
	Port       int      `json:"port" default:"8080"`
	ProjectIds []string `json:"project_ids" default:""`

	AssignedTreatmentLogger AssignedTreatmentLoggerConfig `json:"assigned_treatment_logger"`
	DeploymentConfig        DeploymentConfig              `json:"deployment_config"`
	ManagementService       ManagementServiceConfig       `json:"management_service"`
	MonitoringConfig        Monitoring                    `json:"monitoring_config"`
	SwaggerConfig           SwaggerConfig                 `json:"swagger_config"`
}

type TreatmentServiceConfig struct {
	Port       int      `json:"port" default:"8080"`
	ProjectIds []string `json:"project_ids" default:""`

	AssignedTreatmentLogger AssignedTreatmentLoggerConfig `json:"assigned_treatment_logger"`
	DebugConfig             DebugConfig                   `json:"debug_config"`
	NewRelicConfig          NewRelicConfig                `json:"new_relic_config"`
	SentryConfig            SentryConfig                  `json:"sentry_config"`
	DeploymentConfig        DeploymentConfig              `json:"deployment_config"`
	PubSub                  PubSub                        `json:"pub_sub"`
	ManagementService       ManagementServiceConfig       `json:"management_service"`
	MonitoringConfig        Monitoring                    `json:"monitoring_config"`
	SwaggerConfig           SwaggerConfig                 `json:"swagger_config"`
	SegmenterConfig         map[string]interface{}        `json:"segmenter_config"`
}

type AssignedTreatmentLoggerConfig struct {
	Kind                 string `json:"kind" default:""`
	QueueLength          int    `json:"queue_length" default:"100"`
	FlushIntervalSeconds int    `json:"flush_interval_seconds" default:"1"`

	BQConfig    *BigqueryConfig `json:"bq_config"`
	KafkaConfig *KafkaConfig    `json:"kafka_config"`
}

type DebugConfig struct {
	OutputPath string `json:"output_path" default:"/tmp"`
}

type NewRelicConfig struct {
	Enabled           bool                   `json:"enabled" validate:"required" default:"false"`
	AppName           string                 `json:"app_name"`
	License           string                 `json:"license"`
	Labels            map[string]interface{} `json:"labels"`
	IgnoreStatusCodes []int                  `json:"ignore_status_codes"`
}

type SentryConfig struct {
	Enabled bool              `json:"enabled" validate:"required" default:"false"`
	DSN     string            `json:"dsn"`
	Labels  map[string]string `json:"labels"`
}

type DeploymentConfig struct {
	EnvironmentType string `json:"environment_type" default:"local"`
	MaxGoRoutines   int    `json:"max_go_routines" default:"100"`
}

type PubSub struct {
	Project              string `json:"project" default:"dev"`
	TopicName            string `json:"topic_name" default:"xp-update"`
	PubSubTimeoutSeconds int    `json:"pub_sub_timeout_seconds" default:"30"`
}

type BigqueryConfig struct {
	Project string `json:"project"`
	Dataset string `json:"dataset"`
	Table   string `json:"table"`
}

type KafkaConfig struct {
	Brokers          string `json:"brokers"`
	Topic            string `json:"topics"`
	MaxMessageBytes  int    `json:"max_message_bytes" default:"1048588"`
	CompressionType  string `json:"compression_type" default:"none"`
	ConnectTimeoutMS int    `json:"connect_timeout_ms" default:"1000"`
}

type ManagementServiceConfig struct {
	URL                  string `json:"url" default:"http://localhost:3000/v1"`
	AuthorizationEnabled bool   `json:"authorization_enabled"`
}

type Monitoring struct {
	Kind         string   `json:"kind" default:""`
	MetricLabels []string `json:"metric_labels" default:""`
}

type SwaggerConfig struct {
	Enabled          bool     `json:"enabled" validate:"required" default:"false"`
	AllowedOrigins   []string `json:"allowed_origins" default:"*"`
	OpenAPISpecsPath string   `json:"open_api_specs_path" default:"."`
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
