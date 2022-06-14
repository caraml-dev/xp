package models

import "github.com/gojek/xp/common/api/schema"

type TreatmentHistory struct {
	// CreatedAt - the current value of the UpdatedAt timestamp of the treatment.
	//             This is effectively the time when the version was created.
	// UpdatedAt - the time of creation of the treatment history record.
	// The treatment history record is immutable.
	Model

	// ID is the id of the TreatmentHistory record
	ID ID `json:"id" gorm:"primary_key"`

	// TreatmentID is the id of the treatment whose version this record represents
	TreatmentID ID `json:"treatment_id"`

	// Version is the version number of the treatment, starts at 1 for each treatment.
	Version int64 `json:"version"`

	// The following values are copied from the treatment record at the time of versioning
	Name          string          `json:"name"`
	Configuration TreatmentConfig `json:"configuration"`
	UpdatedBy     string          `json:"updated_by"`
}

// TableName overrides Gorm's default pluralised name: "treatment_histories"
func (TreatmentHistory) TableName() string {
	return "treatment_history"
}

// ToApiSchema converts the treatment history DB model to a format compatible with the
// OpenAPI specifications.
func (t *TreatmentHistory) ToApiSchema() schema.TreatmentHistory {

	return schema.TreatmentHistory{
		Id:            t.ID.ToApiSchema(),
		Name:          t.Name,
		TreatmentId:   t.TreatmentID.ToApiSchema(),
		Configuration: t.Configuration,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		UpdatedBy:     t.UpdatedBy,
		Version:       t.Version,
	}
}
