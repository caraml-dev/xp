package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFloat64Type(t *testing.T) {

	tests := map[string]struct {
		value    interface{}
		key      string
		expected error
	}{
		"success": {
			value:    1.23,
			key:      "longitude",
			expected: nil,
		},
		"failed | incorrect type": {
			value:    "1.23",
			key:      "longitude",
			expected: fmt.Errorf("incorrect type provided for longitude; expected float64"),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := ValidateFloat64Type(data.value, data.key)

			assert.Equal(t, data.expected, resp)
		})
	}
}
