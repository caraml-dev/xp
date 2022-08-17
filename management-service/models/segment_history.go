package models

import "github.com/caraml-dev/xp/common/api/schema"

type SegmentHistory struct {
	// CreatedAt - the current value of the UpdatedAt timestamp of the segment.
	//             This is effectively the time when the version was created.
	// UpdatedAt - the time of creation of the segment history record.
	// The segment history record is immutable.
	Model

	// ID is the id of the SegmentHistory record
	ID ID `json:"id" gorm:"primary_key"`

	// SegmentID is the id of the segment whose version this record represents
	SegmentID ID `json:"segment_id"`

	// Version is the version number of the segment, starts at 1 for each segment.
	Version int64 `json:"version"`

	// The following values are copied from the segment record at the time of versioning
	Name      string            `json:"name"`
	Segment   ExperimentSegment `json:"segment"`
	UpdatedBy string            `json:"updated_by"`
}

// TableName overrides Gorm's default pluralised name: "segment_histories"
func (SegmentHistory) TableName() string {
	return "segment_history"
}

// ToApiSchema converts the segment history DB model to a format compatible with the
// OpenAPI specifications.
func (s *SegmentHistory) ToApiSchema(segmentersType map[string]schema.SegmenterType) schema.SegmentHistory {

	return schema.SegmentHistory{
		Id:        s.ID.ToApiSchema(),
		Name:      s.Name,
		SegmentId: s.SegmentID.ToApiSchema(),
		Segment:   s.Segment.ToApiSchema(segmentersType),
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		UpdatedBy: s.UpdatedBy,
		Version:   s.Version,
	}
}
