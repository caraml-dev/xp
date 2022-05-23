package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/xp/common/api/schema"
)

func TestTreatmentHistoryToApiSchema(t *testing.T) {
	name := "treatment-hist-1"
	updatedBy := "test-updated-by"
	config := map[string]interface{}{
		"config-1": "value",
		"config-2": 2,
	}
	treatment := TreatmentHistory{
		Model: Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:            ID(100),
		TreatmentID:   ID(40),
		Version:       int64(8),
		Name:          name,
		Configuration: config,
		UpdatedBy:     updatedBy,
	}
	assert.Equal(t, schema.TreatmentHistory{
		Id:            int64(100),
		CreatedAt:     time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		TreatmentId:   int64(40),
		Version:       int64(8),
		Name:          name,
		Configuration: config,
		UpdatedBy:     updatedBy,
	}, treatment.ToApiSchema())
}
