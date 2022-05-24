package util

import (
	"errors"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
)

func TestGetS2ID(t *testing.T) {
	tests := map[string]struct {
		lat      float64
		long     float64
		level    int
		err      error
		expected s2.CellID
	}{
		"success": {
			long:     103.8998991137485,
			lat:      1.2537040223936706,
			level:    14,
			expected: s2.CellID(3592210809859604480),
		},
		"failure | incorrect lat long": {
			long:     100000,
			lat:      100000,
			level:    14,
			err:      errors.New("received invalid latitude, longitude values"),
			expected: s2.CellID(1),
		},
		"failure | invalid level": {
			long:     103.8998991137485,
			lat:      1.2537040223936706,
			level:    50,
			err:      errors.New("received invalid s2 geo level"),
			expected: s2.CellID(1),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := GetS2ID(data.lat, data.long, data.level)

			if data.err != nil {
				assert.Equal(t, data.err, err)
			}

			if data.err == nil {
				assert.Equal(t, data.expected, resp)
			}
		})
	}
}
