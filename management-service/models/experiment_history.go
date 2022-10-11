package models

import (
	"time"

	"github.com/caraml-dev/xp/common/api/schema"
	"gorm.io/gorm"
)

type ExperimentHistory struct {
	// CreatedAt - the current value of the UpdatedAt timestamp of the experiment.
	//             This is effectively the time when the version was created.
	// UpdatedAt - the time of creation of the experiment history record.
	// The experiment history record is immutable.
	Model

	// ID is the id of the ExperimentHistory record
	ID ID `json:"id" gorm:"primary_key"`

	// ExperimentID is the id of the experiment whose version this record represents
	ExperimentID ID `json:"experiment_id"`

	// Version is the version number of the experiment, starts at 1 for each experiment.
	Version int64 `json:"version"`

	// The following values are copied from the experiment record at the time of versioning
	Name        string               `json:"name"`
	Description *string              `json:"description"`
	Type        ExperimentType       `json:"type"`
	Interval    *int32               `json:"interval"`
	Tier        ExperimentTier       `json:"tier"`
	Treatments  ExperimentTreatments `json:"treatments"`
	Segment     ExperimentSegment    `json:"segment"`
	Status      ExperimentStatus     `json:"status"`
	StartTime   time.Time            `json:"start_time"`
	EndTime     time.Time            `json:"end_time"`
	UpdatedBy   string               `json:"updated_by"`
}

// TableName overrides Gorm's default pluralised name: "experiment_histories"
func (ExperimentHistory) TableName() string {
	return "experiment_history"
}

func (e *ExperimentHistory) AfterFind(tx *gorm.DB) error {
	e.StartTime = e.StartTime.In(utcLoc)
	e.EndTime = e.EndTime.In(utcLoc)
	return nil
}

// ToApiSchema converts the experiment history DB model to a format compatible with the
// OpenAPI specifications.
func (e *ExperimentHistory) ToApiSchema(segmentersType map[string]schema.SegmenterType) schema.ExperimentHistory {
	status := schema.ExperimentStatus(e.Status)
	expType := schema.ExperimentType(e.Type)
	tierType := schema.ExperimentTier(e.Tier)

	return schema.ExperimentHistory{
		Description:  e.Description,
		EndTime:      e.EndTime,
		Id:           e.ID.ToApiSchema(),
		Interval:     e.Interval,
		Name:         e.Name,
		ExperimentId: e.ExperimentID.ToApiSchema(),
		Segment:      e.Segment.ToApiSchema(segmentersType),
		Status:       status,
		Tier:         tierType,
		Treatments:   e.Treatments.ToApiSchema(),
		Type:         expType,
		StartTime:    e.StartTime,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		UpdatedBy:    e.UpdatedBy,
		Version:      e.Version,
	}
}
