package models

import "time"

var utcLoc, _ = time.LoadLocation("UTC")

type ID uint

func (id ID) ToApiSchema() int64 {
	return int64(id)
}

// Model is a struct containing the basic fields for a persisted entity defined
// in the API.
type Model struct {
	// Created timestamp. Populated when the object is saved to the db.
	CreatedAt time.Time `json:"created_at"`
	// Last updated timestamp. Updated when the object is updated in the db.
	UpdatedAt time.Time `json:"updated_at"`
}

func (m Model) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m Model) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}
