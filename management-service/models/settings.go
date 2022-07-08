package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/xp/common/api/schema"
	_pubsub "github.com/gojek/xp/common/pubsub"
)

type ProjectSegmenters struct {
	Names     []string            `json:"names"`
	Variables map[string][]string `json:"variables"`
}

type ExperimentationConfig struct {
	// Segmenters is a list of names of segmenters chosen for the project
	Segmenters ProjectSegmenters `json:"segmenters"`
	// RandomizationKey is the of the randomization key in the request payload
	RandomizationKey string `json:"randomization_key"`
	// S2IDClusteringEnabled determines whether S2ID cluster ID should be used
	// as the randomization key, for randomized switchback experiments
	S2IDClusteringEnabled bool `json:"enable_s2id_clustering"`
}

type Rule struct {
	// Name is the name of the rule
	Name string `json:"name" validate:"required,notBlank"`
	// Predicate is the predicate of the rule
	Predicate string `json:"predicate" validate:"required,notBlank"`
}

type TreatmentSchema struct {
	Rules []Rule `json:"rules" validate:"required,unique=Name,dive,required"`
}

// Settings stores the project's Experimentation settings
type Settings struct {
	Model

	// ProjectID is the id of the MLP project
	ProjectID ID `json:"project_id" gorm:"primary_key;auto_increment:false"`

	// Username is used for authentication by the Fetch Treatment API
	Username string `json:"username"`
	// Passkey is in plaintext and is used for authentication by the Fetch Treatment API
	Passkey string `json:"passkey"`
	// Config holds the project-wide experimentation configs, as configured by the user
	Config *ExperimentationConfig `json:"config"`
	// TreatmentSchema holds the rules that define the treatment schema
	TreatmentSchema *TreatmentSchema `json:"treatment_schema"`
	// ValidationUrl holds the custom validation endpoint defined by the user
	ValidationUrl *string `json:"validation_url"`
}

func (ec *ExperimentationConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &ec)
}

func (ec ExperimentationConfig) Value() (driver.Value, error) {
	return json.Marshal(ec)
}

func (ts *TreatmentSchema) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &ts)
}

func (ts TreatmentSchema) Value() (driver.Value, error) {
	return json.Marshal(ts)
}

func (ts *TreatmentSchema) ToOpenApi() *schema.TreatmentSchema {
	if ts == nil {
		return nil
	}

	var treatmentSchemaRules schema.Rules
	for _, rule := range ts.Rules {
		treatmentSchemaRules = append(treatmentSchemaRules, schema.Rule{Name: rule.Name, Predicate: rule.Predicate})
	}
	return &schema.TreatmentSchema{Rules: treatmentSchemaRules}
}

// ToApiSchema converts the settings DB model to a format compatible with the
// OpenAPI specifications.
func (c *Settings) ToApiSchema() schema.ProjectSettings {

	user := schema.ProjectSettings{
		CreatedAt:            c.CreatedAt,
		EnableS2idClustering: c.Config.S2IDClusteringEnabled,
		Passkey:              c.Passkey,
		ProjectId:            c.ProjectID.ToApiSchema(),
		RandomizationKey:     c.Config.RandomizationKey,
		Segmenters: schema.ProjectSegmenters{
			Names:     c.Config.Segmenters.Names,
			Variables: schema.ProjectSegmenters_Variables{AdditionalProperties: c.Config.Segmenters.Variables},
		},
		UpdatedAt:       c.UpdatedAt,
		Username:        c.Username,
		TreatmentSchema: c.TreatmentSchema.ToOpenApi(),
		ValidationUrl:   c.ValidationUrl,
	}

	return user
}

func (c *Settings) ToProtoSchema() _pubsub.ProjectSettings {

	segmentersVariables := make(map[string]*_pubsub.ExperimentVariables)
	for segmenterName, experimentVariables := range c.Config.Segmenters.Variables {
		protoExperimentVariables := _pubsub.ExperimentVariables{
			Value: experimentVariables,
		}
		segmentersVariables[segmenterName] = &protoExperimentVariables
	}

	projectSegmenters := _pubsub.Segmenters{
		Names:     c.Config.Segmenters.Names,
		Variables: segmentersVariables,
	}

	return _pubsub.ProjectSettings{
		ProjectId:            c.ProjectID.ToApiSchema(),
		CreatedAt:            timestamppb.New(c.CreatedAt),
		UpdatedAt:            timestamppb.New(c.UpdatedAt),
		Username:             c.Username,
		Passkey:              c.Passkey,
		EnableS2IdClustering: c.Config.S2IDClusteringEnabled,
		Segmenters:           &projectSegmenters,
		RandomizationKey:     c.Config.RandomizationKey,
	}
}
