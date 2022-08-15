package config

import (
	"sync"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
)

// RequestParameter captures a single parameter's config for parsing the incoming
// request and extracting values
type RequestParameter struct {
	// Parameter specifies the name of the parameter, expected by the XP API
	Parameter string `json:"parameter" validate:"required"`
	// Field specifies the name of the field in the incoming request header/payload
	Field string `json:"field" validate:"required"`
	// FieldSrc specifies whether the field is located in the request header or payload
	FieldSrc request.FieldSource `json:"field_source" validate:"required,oneof=header payload"`
}

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
	Enabled        bool           `json:"enabled"`
	BaseURL        string         `json:"base_url"`      // Base URL for XP experiment REST API
	HomePageURL    string         `json:"home_page_url"` // Website URL for end-users to manage experiments
	RemoteUI       RemoteUI       `json:"remote_ui"`
	RunnerDefaults RunnerDefaults `json:"runner_defaults"`
}

// ExperimentRunnerConfig is used to parse the XP runner config during initialization
type ExperimentRunnerConfig struct {
	Endpoint          string             `json:"endpoint" validate:"required"`
	ProjectID         int                `json:"project_id" validate:"required"`
	Passkey           string             `json:"passkey" validate:"required"`
	Timeout           string             `json:"timeout" validate:"required"`
	RequestParameters []RequestParameter `json:"request_parameters" validate:"required,dive"`
}

type Variable struct {
	Name        string              `json:"name" validate:"required"`
	Field       string              `json:"field" validate:"required"`
	FieldSource request.FieldSource `json:"field_source" validate:"required,oneof=header payload"`
}

// ExperimentConfig is the experiment config saved on the Turing DB
type ExperimentConfig struct {
	// Keeping the ProjectID as a separate parameter for now, in the chance that
	// we need flexibility in the router -> experiment project association.
	ProjectID int        `json:"project_id"  validate:"required"`
	Variables []Variable `json:"variables" validate:"dive"`
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
		v := validator.New()
		// Save the validator to the global state
		experimentRunnerConfigValidator = v
	})

	return experimentRunnerConfigValidator
}
