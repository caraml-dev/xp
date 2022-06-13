package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/turing-experiments/common/api/schema"
)

func TestProjectToApiSchema(t *testing.T) {
	createdUpdatedAt := time.Now()
	id := int64(0)
	randomizationKey := "random"
	segmenters := []string{"seg-1", "seg-2"}
	username := "user1"
	testProject := Project{
		CreatedAt:        createdUpdatedAt,
		Id:               id,
		RandomizationKey: randomizationKey,
		Segmenters:       segmenters,
		UpdatedAt:        createdUpdatedAt,
		Username:         username,
	}

	assert.Equal(t, schema.Project{
		CreatedAt:        createdUpdatedAt,
		Id:               0,
		RandomizationKey: randomizationKey,
		Segmenters:       segmenters,
		UpdatedAt:        createdUpdatedAt,
		Username:         username,
	}, testProject.ToApiSchema())
}
