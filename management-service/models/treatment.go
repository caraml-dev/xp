package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/caraml-dev/xp/common/api/schema"
)

type TreatmentField string

// Defines values for TreatmentField.
const (
	TreatmentFieldId TreatmentField = "id"

	TreatmentFieldName TreatmentField = "name"
)

type TreatmentConfig map[string]interface{}

func (t *TreatmentConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &t)
}

func (t TreatmentConfig) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type Treatment struct {
	Model

	// ID is the id of the Treatment
	ID ID `json:"id" gorm:"primary_key"`

	// ProjectID is the id of the project that this client belongs to,
	// as retrieved from the MLP API.
	ProjectID ID `json:"project_id"`

	// Name is the treatment's name
	Name string `json:"name"`

	Configuration TreatmentConfig `json:"configuration"`

	// UpdatedBy holds the details of the last person/job that updated the experiment
	UpdatedBy string `json:"updated_by"`
}

// ToApiSchema converts the configured treatment DB model to a format compatible with the
// OpenAPI specifications.
func (t *Treatment) ToApiSchema(fields ...TreatmentField) schema.Treatment {
	treatment := schema.Treatment{}
	// Only return requested fields
	if fields != nil {
		for _, field := range fields {
			switch field {
			case "name":
				treatment.Name = &t.Name
			case "id":
				id := t.ID.ToApiSchema()
				treatment.Id = &id
			}
		}
		return treatment
	}

	id := t.ID.ToApiSchema()
	projectId := t.ProjectID.ToApiSchema()
	config := map[string]interface{}{}
	if t.Configuration != nil {
		config = t.Configuration
	}

	return schema.Treatment{
		Configuration: &config,
		CreatedAt:     &t.CreatedAt,
		Id:            &id,
		Name:          &t.Name,
		ProjectId:     &projectId,
		UpdatedAt:     &t.UpdatedAt,
		UpdatedBy:     &t.UpdatedBy,
	}

}
