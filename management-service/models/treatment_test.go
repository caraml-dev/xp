package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/turing-experiments/common/api/schema"
)

var testTreatment = Treatment{
	Model: Model{
		CreatedAt: time.Date(2021, 12, 1, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2021, 12, 1, 3, 4, 5, 0, time.UTC),
	},
	ID:        0,
	ProjectID: 1,
	Name:      "test-treatment",
	Configuration: map[string]interface{}{
		"config-1": "value",
		"config-2": 100.5,
		"config-3": map[string]interface{}{
			"x": "y",
			"z": nil,
		},
		"config-4": []interface{}{true, false, true},
	},
	UpdatedBy: "test-user",
}

func TestTreatmentToApiSchema(t *testing.T) {
	id := int64(0)
	projectId := int64(1)
	name := "test-treatment"
	createdAt := time.Date(2021, 12, 1, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2021, 12, 1, 3, 4, 5, 0, time.UTC)
	updatedBy := "test-user"
	assert.Equal(t, schema.Treatment{
		Id:        &id,
		ProjectId: &projectId,
		Configuration: &map[string]interface{}{
			"config-1": "value",
			"config-2": 100.5,
			"config-3": map[string]interface{}{
				"x": "y",
				"z": nil,
			},
			"config-4": []interface{}{true, false, true},
		},
		Name:      &name,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, testTreatment.ToApiSchema())
}

func TestTreatmentToApiSchemaWithFields(t *testing.T) {
	id := int64(0)
	name := "test-treatment"

	fields := []TreatmentField{TreatmentFieldId, TreatmentFieldName}
	assert.Equal(t, schema.Treatment{
		Id:   &id,
		Name: &name,
	}, testTreatment.ToApiSchema(fields...))
}
