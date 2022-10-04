package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
)

type ExperimentStatus string
type ExperimentType string
type ExperimentTier string

const (
	ExperimentStatusActive ExperimentStatus = "active"

	ExperimentStatusInactive ExperimentStatus = "inactive"
)

// Defines values for ExperimentType.
const (
	ExperimentTypeAB ExperimentType = "A/B"

	ExperimentTypeSwitchback ExperimentType = "Switchback"
)

// Defines values for ExperimentTier.
const (
	ExperimentTierDefault ExperimentTier = "default"

	ExperimentTierOverride ExperimentTier = "override"
)

type Experiment struct {
	Model

	// ID is the id of the Experiment
	ID ID `json:"id" gorm:"primary_key"`

	// ProjectID is the id of the project that this client belongs to,
	// as retrieved from the MLP API.
	ProjectID ID `json:"project_id"`

	// Version is the version number of the experiment, starts at 1 for each experiment.
	Version int64 `json:"version"`

	// Name is the experiment's name
	Name string `json:"name"`
	// Description is an optional value that has additional info on the experiment
	Description *string `json:"description"`
	// Type captures the experiment's type
	Type ExperimentType `json:"type"`
	// Interval holds the switchback interval in minutes
	Interval *int32 `json:"interval"`
	// Tier holds the priority of the experiment
	Tier ExperimentTier `json:"tier"`
	// Treatments holds the experiment treatment configurations
	Treatments ExperimentTreatments `json:"treatments"`
	// Segment holds the combination of segmenters that the experiment applies to
	Segment ExperimentSegment `json:"segment"`
	// Status is the experiment's status
	Status ExperimentStatus `json:"status"`
	// StartTime describes the time at which an experiment starts
	StartTime time.Time `json:"start_time"`
	// EndTime describes the time at which an experiment ends
	EndTime time.Time `json:"end_time"`
	// UpdatedBy holds the details of the last person/job that updated the experiment
	UpdatedBy string `json:"updated_by"`
}

// ToApiSchema converts the experiment DB model to a format compatible with the
// OpenAPI specifications.
func (e *Experiment) ToApiSchema(segmentersType map[string]schema.SegmenterType) schema.Experiment {
	return schema.Experiment{
		Description:    e.Description,
		EndTime:        e.EndTime,
		Id:             e.ID.ToApiSchema(),
		Interval:       e.Interval,
		Name:           e.Name,
		ProjectId:      e.ProjectID.ToApiSchema(),
		Segment:        e.Segment.ToApiSchema(segmentersType),
		Status:         schema.ExperimentStatus(e.Status),
		StatusFriendly: getExperimentStatusFriendly(e.StartTime, e.EndTime, e.Status),
		Treatments:     e.Treatments.ToApiSchema(),
		Type:           schema.ExperimentType(e.Type),
		Tier:           schema.ExperimentTier(e.Tier),
		StartTime:      e.StartTime,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		UpdatedBy:      e.UpdatedBy,
		Version:        e.Version,
	}
}

// ToProtoSchema converts the experiment DB model to a format compatible with the
// Protobuf specifications.
func (e *Experiment) ToProtoSchema(segmentersType map[string]schema.SegmenterType) (*_pubsub.Experiment, error) {
	var interval int32
	if e.Interval != nil {
		interval = *e.Interval
	}

	var experimentStatus _pubsub.Experiment_Status
	switch e.Status {
	case ExperimentStatusActive:
		experimentStatus = _pubsub.Experiment_Active
	case ExperimentStatusInactive:
		experimentStatus = _pubsub.Experiment_Inactive
	}

	var experimentTier _pubsub.Experiment_Tier
	switch e.Tier {
	case ExperimentTierDefault:
		experimentTier = _pubsub.Experiment_Default
	case ExperimentTierOverride:
		experimentTier = _pubsub.Experiment_Override
	}

	var experimentType _pubsub.Experiment_Type
	switch e.Type {
	case ExperimentTypeSwitchback:
		experimentType = _pubsub.Experiment_Switchback
	case ExperimentTypeAB:
		experimentType = _pubsub.Experiment_A_B
	}

	segments := e.Segment.ToProtoSchema(segmentersType)
	treatments, err := e.Treatments.ToProtoSchema()
	if err != nil {
		return nil, err
	}

	startTime := timestamppb.New(e.StartTime)
	endTime := timestamppb.New(e.EndTime)
	updatedAt := timestamppb.New(e.UpdatedAt)

	return &_pubsub.Experiment{
		ProjectId:  e.ProjectID.ToApiSchema(),
		EndTime:    endTime,
		Id:         e.ID.ToApiSchema(),
		Interval:   interval,
		Name:       e.Name,
		Segments:   segments,
		Status:     experimentStatus,
		Treatments: treatments,
		Tier:       experimentTier,
		Type:       experimentType,
		StartTime:  startTime,
		UpdatedAt:  updatedAt,
		Version:    e.Version,
	}, nil
}

func getExperimentStatusFriendly(startTime time.Time, endTime time.Time, status ExperimentStatus) schema.ExperimentStatusFriendly {
	statusFriendly := schema.ExperimentStatusFriendlyDeactivated
	if status == ExperimentStatusActive {
		currentTime := time.Now()
		if currentTime.Before(startTime) {
			statusFriendly = schema.ExperimentStatusFriendlyScheduled
		} else if currentTime.After(endTime) {
			statusFriendly = schema.ExperimentStatusFriendlyCompleted
		} else {
			statusFriendly = schema.ExperimentStatusFriendlyRunning
		}
	}
	return statusFriendly
}
