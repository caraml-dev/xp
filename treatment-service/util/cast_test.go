package util

import (
	"hash/fnv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFloat64ToInt64(t *testing.T) {
	val := float64(1)
	resp, _ := ConvertFloat64ToInt64(val)
	expected := int64(1)

	assert.Equal(t, expected, resp)
}

func TestHash(t *testing.T) {
	val := "test"
	h := fnv.New32a()
	h.Write([]byte(val))
	expected := h.Sum32()

	resp := Hash(val)

	assert.Equal(t, expected, resp)
}

func TestDereferenceString(t *testing.T) {
	refString := "test123"
	tests := map[string]struct {
		ref      *string
		default_ string
		expected string
	}{
		"success | default": {
			ref:      nil,
			default_: "",
			expected: "",
		},
		"success | ref": {
			ref:      &refString,
			default_: "",
			expected: refString,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := DereferenceString(data.ref, data.default_)

			assert.Equal(t, data.expected, resp)
		})
	}
}

func TestDereferenceInt(t *testing.T) {
	refInt := int64(1)
	tests := map[string]struct {
		ref      *int64
		default_ int64
		expected int64
	}{
		"success | default": {
			ref:      nil,
			default_: 2,
			expected: 2,
		},
		"success | ref": {
			ref:      &refInt,
			default_: 2,
			expected: refInt,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := DereferenceInt(data.ref, data.default_)

			assert.Equal(t, data.expected, resp)
		})
	}
}

func TestDereferenceUInt(t *testing.T) {
	defaultUInt := uint64(2)
	refUInt := uint64(1)

	tests := map[string]struct {
		ref      *uint64
		default_ uint64
		expected uint64
	}{
		"success | default": {
			ref:      nil,
			default_: defaultUInt,
			expected: defaultUInt,
		},
		"success | ref": {
			ref:      &refUInt,
			default_: defaultUInt,
			expected: refUInt,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := DereferenceUInt(data.ref, data.default_)

			assert.Equal(t, data.expected, resp)
		})
	}
}

func TestDereferenceBool(t *testing.T) {
	defaultBool := false
	refBool := true

	tests := map[string]struct {
		ref      *bool
		default_ bool
		expected bool
	}{
		"success | default": {
			ref:      nil,
			default_: defaultBool,
			expected: defaultBool,
		},
		"success | ref": {
			ref:      &refBool,
			default_: defaultBool,
			expected: refBool,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := DereferenceBool(data.ref, data.default_)

			assert.Equal(t, data.expected, resp)
		})
	}
}

func TestGetFloatSegmenter(t *testing.T) {
	tests := map[string]struct {
		values    map[string]interface{}
		key       string
		segmenter string
		expected  float64
		errString string
	}{
		"failure | bool": {
			values:    map[string]interface{}{"key": true},
			key:       "key",
			segmenter: "float-seg",
			errString: "invalid type of variable (key) was provided for float-seg segmenter; expected float64",
		},
		"success | float64": {
			values:    map[string]interface{}{"key": 1.2537040223936706},
			key:       "key",
			segmenter: "float-seg",
			expected:  1.2537040223936706,
		},
		"success | string": {
			values:    map[string]interface{}{"key": "1.2537040223936706"},
			key:       "key",
			segmenter: "float-seg",
			expected:  1.2537040223936706,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := GetFloatSegmenter(data.values, data.key, data.segmenter)

			if data.errString == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, *resp)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}
