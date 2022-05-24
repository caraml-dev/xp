package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModelGetters(t *testing.T) {
	model := Model{
		CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt: time.Date(2020, 1, 1, 2, 3, 4, 0, time.UTC),
	}

	assert.Equal(t, time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC), model.GetCreatedAt())
	assert.Equal(t, time.Date(2020, 1, 1, 2, 3, 4, 0, time.UTC), model.GetUpdatedAt())
}

func TestIDToApiSchema(t *testing.T) {
	assert.Equal(t, int64(40), ID(40).ToApiSchema())
}
