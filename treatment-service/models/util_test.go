package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsProjectId(t *testing.T) {
	tests := map[string]struct {
		actual        []ProjectId
		expectedValue ProjectId
		expected      bool
	}{
		"success | present": {
			actual:        []ProjectId{ProjectId(1), ProjectId(2)},
			expectedValue: ProjectId(1),
			expected:      true,
		},
		"success | not found": {
			actual:        []ProjectId{ProjectId(3)},
			expectedValue: ProjectId(1),
			expected:      false,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := ContainsProjectId(data.actual, data.expectedValue)

			assert.Equal(t, data.expected, resp)
		})
	}
}
