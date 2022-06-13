package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"

	"github.com/gojek/turing-experiments/management-service/config"
	"github.com/gojek/turing-experiments/management-service/errors"
	"github.com/gojek/turing-experiments/management-service/models"
)

var nameRegex = regexp.MustCompile(`^[A-Za-z\d][\w\d \-()#$%&:.]{2,62}[\w\d\-()#$%&:.]$`)

type ValidationService interface {
	Validate(data interface{}) error
	ValidateEntityWithExternalUrl(operation OperationType, entityType EntityType, data interface{}, context ValidationContext,
		validationUrl *string) error
	ValidateWithExternalUrl(reqBody []byte, validationUrl *string) error
}

type validationService struct {
	config                   config.ValidationConfig
	v                        *validator.Validate
	externalValidationClient http.Client
}

func (v *validationService) Validate(data interface{}) error {
	return v.v.Struct(data)
}

// NewValidationService creates a new validator
func NewValidationService(config config.ValidationConfig) (ValidationService, error) {
	instance := validator.New()

	// Register custom validators
	if err := instance.RegisterValidation("notBlank", validators.NotBlank); err != nil {
		return nil, err
	}
	instance.RegisterStructValidation(validateCreateExperimentData, CreateExperimentRequestBody{})
	instance.RegisterStructValidation(validateUpdateExperimentData, UpdateExperimentRequestBody{})
	instance.RegisterStructValidation(validateCreateTreatmentData, CreateTreatmentRequestBody{})
	instance.RegisterStructValidation(validateCreateProjectSettingsData, CreateProjectSettingsRequestBody{})
	instance.RegisterStructValidation(validateUpdateProjectSettingsData, UpdateProjectSettingsRequestBody{})

	externalValidationClient := http.Client{
		Timeout: time.Duration(config.ValidationUrlTimeoutSeconds) * time.Second,
	}
	return &validationService{config: config, v: instance, externalValidationClient: externalValidationClient}, nil
}

func validateCreateExperimentData(sl validator.StructLevel) {
	field := sl.Current().Interface().(CreateExperimentRequestBody)
	checkName(sl, "Name", field.Name)
	checkStartTime(sl, field.StartTime)
	checkInterval(sl, field.Type, field.Interval)
	checkTreatments(sl, field.Type, field.Treatments)
}

func validateUpdateExperimentData(sl validator.StructLevel) {
	field := sl.Current().Interface().(UpdateExperimentRequestBody)
	checkStartTime(sl, field.StartTime)
	checkInterval(sl, field.Type, field.Interval)
	checkTreatments(sl, field.Type, field.Treatments)
}

func validateCreateTreatmentData(sl validator.StructLevel) {
	field := sl.Current().Interface().(CreateTreatmentRequestBody)
	checkName(sl, "Name", field.Name)
}

func validateCreateProjectSettingsData(sl validator.StructLevel) {
	field := sl.Current().Interface().(CreateProjectSettingsRequestBody)
	checkTreatmentSchema(sl, field.TreatmentSchema)
}

func validateUpdateProjectSettingsData(sl validator.StructLevel) {
	field := sl.Current().Interface().(UpdateProjectSettingsRequestBody)
	checkTreatmentSchema(sl, field.TreatmentSchema)
}

func checkName(sl validator.StructLevel, fieldName string, value string) {
	nameRegexDescription := strings.Join([]string{
		"Name must be between 4-64 characters long, and begin with an alphanumeric character",
		"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.",
	}, " ")
	if !nameRegex.MatchString(value) {
		sl.ReportError(value, fieldName, "name", nameRegexDescription, fmt.Sprintf("%v", value))
	}
}

func checkStartTime(sl validator.StructLevel, startTime time.Time) {
	if startTime.Before(time.Now()) {
		sl.ReportError(startTime, "StartTime", "start_time", "start-time-in-future", fmt.Sprintf("%v", startTime))
	}
}

func checkInterval(sl validator.StructLevel, experimentType models.ExperimentType, interval *int32) {
	switch experimentType {
	case models.ExperimentTypeAB:
		// Interval should not be set for a/b experiment
		if interval != nil {
			sl.ReportError(interval, "Interval", "interval", "interval-unset-ab-experiment", fmt.Sprintf("%d", *interval))
		}
	case models.ExperimentTypeSwitchback:
		// Interval should be set for switchback experiment
		if interval == nil || *interval <= 0 {
			sl.ReportError(0, "Interval", "interval", "interval-set-switchback-experiment", "")
		}
	}
}

func checkTreatments(sl validator.StructLevel, experimentType models.ExperimentType, treatments models.ExperimentTreatments) {
	// This needs to be checked here because the OpenAPI tag generation does not work for arrays
	err := sl.Validator().Var(treatments, "notBlank")
	if err != nil {
		sl.ReportError(treatments, "Treatments", "treatments", "notBlank", fmt.Sprintf("%v", treatments))
	}

	// Check treatment names
	for _, treatment := range treatments {
		checkName(sl, "Treatments", treatment.Name)
	}

	// Check that the traffic sum is 100 for AB and 0 or 100 for Switchback
	trafficSum := int32(0)
	for _, treatment := range treatments {
		if treatment.Traffic != nil {
			trafficSum += *treatment.Traffic
		}
	}

	// If traffic sum is non-zero, there should be no treatments with 0 traffic.
	if trafficSum != 0 {
		for _, treatment := range treatments {
			if treatment.Traffic != nil && *treatment.Traffic == 0 {
				sl.ReportError(treatments, "Treatments", "treatments", "traffic-is-0", fmt.Sprintf("%d", *treatment.Traffic))
			}
		}
	}
	switch experimentType {
	case models.ExperimentTypeAB:
		// Traffic should add to 100
		if trafficSum != 100 {
			sl.ReportError(treatments, "Treatments", "treatments", "traffic-sum-100", fmt.Sprintf("%d", trafficSum))
		}
	case models.ExperimentTypeSwitchback:
		// Switchback experiments can either have no traffic defined (cyclic switchback),
		// or the traffic should add to 100 (randomised switchback)
		if trafficSum != 0 && trafficSum != 100 {
			sl.ReportError(treatments, "Treatments", "treatments", "traffic-sum-0-or-100", fmt.Sprintf("%d", trafficSum))
		}
	}
}

func checkTreatmentSchema(sl validator.StructLevel, treatmentSchema *models.TreatmentSchema) {
	if treatmentSchema == nil {
		return
	}
	for _, rule := range treatmentSchema.Rules {
		if err := CheckRulePredicate(rule.Predicate); err != nil {
			sl.ReportError(treatmentSchema, "TreatmentSchema", "treatmentSchema", fmt.Sprintf("invalid-predicate: %v", err),
				fmt.Sprintf("%v", rule.Name))
		}
	}
}

// CheckRulePredicate checks if a given predicate is a valid Go template expression
func CheckRulePredicate(predicate string) error {
	_, err := template.New("").Funcs(sprig.FuncMap()).Parse(predicate)

	if err != nil {
		return err
	}
	return nil
}

type EntityType string

const (
	EntityTypeExperiment EntityType = "experiment"
	EntityTypeTreatment  EntityType = "treatment"
	EntityTypeSegment    EntityType = "segment"
)

type OperationType string

const (
	OperationTypeCreate OperationType = "create"
	OperationTypeUpdate OperationType = "update"
)

type ValidationContext struct {
	CurrentData interface{} `json:"current_data"`
}

type validationUrlRequest struct {
	EntityType    EntityType        `json:"entity_type"`
	OperationType OperationType     `json:"operation"`
	Data          interface{}       `json:"data"`
	Context       ValidationContext `json:"context"`
}

// ValidateEntityWithExternalUrl validates the given entity by sending it as part of the payload, together with
// information on its context, operation and entity type, to a user-defined HTTP endpoint; if the response from the
// endpoint is anything but a 200, this method will return an error
func (v *validationService) ValidateEntityWithExternalUrl(
	operation OperationType,
	entityType EntityType,
	data interface{},
	context ValidationContext,
	validationUrl *string,
) error {
	if validationUrl == nil {
		return nil
	}

	validationRequest := validationUrlRequest{
		EntityType:    entityType,
		OperationType: operation,
		Data:          data,
		Context:       context,
	}
	reqBody, err := json.Marshal(validationRequest)
	if err != nil {
		return errors.Newf(errors.BadInput, "Error marshalling the validation request: %v", err.Error())
	}

	err = v.ValidateWithExternalUrl(reqBody, validationUrl)
	if err != nil {
		return err
	}

	return nil
}

// ValidateWithExternalUrl validates the given payload request by sending it to a user-defined HTTP endpoint; if the
// response from the endpoint is anything but a 200, this method will return an error
func (v *validationService) ValidateWithExternalUrl(
	reqBody []byte,
	validationUrl *string,
) error {
	req, err := http.NewRequest("POST", *validationUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return errors.Newf(errors.BadInput, "Error creating the HTTP request: %v", err.Error())
	}

	resp, err := v.externalValidationClient.Do(req)
	if err != nil {
		return errors.Newf(errors.BadInput, "Error sending request to custom validation endpoint: %v", err.Error())
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf(errors.BadInput, "Error validating data with validation URL: %v", resp.Status)
	}

	return nil
}
