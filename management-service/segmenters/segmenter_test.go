package segmenters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type SegmenterTestSuite struct {
	suite.Suite
	SegmenterConfigs []*_segmenters.SegmenterConfiguration
}

func (s *SegmenterTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmenterTestSuite")

	s.SegmenterConfigs = []*_segmenters.SegmenterConfiguration{
		{
			Name: "test-segmenter-1",
			Type: _segmenters.SegmenterValueType_INTEGER,
			Options: map[string]*_segmenters.SegmenterValue{
				"zero": {Value: &_segmenters.SegmenterValue_Integer{Integer: 0}},
				"one":  {Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
				"two":  {Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
			},
			MultiValued: true,
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"test-segmenter-1"},
					},
				},
			},
		},
		{
			Name: "test-segmenter-2",
			Type: _segmenters.SegmenterValueType_STRING,
			Options: map[string]*_segmenters.SegmenterValue{
				"zero": {Value: &_segmenters.SegmenterValue_String_{String_: "0"}},
				"one":  {Value: &_segmenters.SegmenterValue_String_{String_: "1"}},
				"two":  {Value: &_segmenters.SegmenterValue_String_{String_: "2"}},
			},
			MultiValued: false,
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"exp-var-1", "exp-var-2"},
					},
				},
			},
		},
		{
			Name:        "test-segmenter-3",
			Type:        _segmenters.SegmenterValueType_REAL,
			Options:     map[string]*_segmenters.SegmenterValue{},
			MultiValued: false,
		},
		{
			Name:        "test-segmenter-4",
			Type:        _segmenters.SegmenterValueType_REAL,
			Options:     map[string]*_segmenters.SegmenterValue{},
			MultiValued: true,
			Constraints: []*_segmenters.Constraint{
				{
					PreRequisites: []*_segmenters.PreRequisite{
						{
							SegmenterName: "test-segmenter-1",
							SegmenterValues: &_segmenters.ListSegmenterValue{
								Values: []*_segmenters.SegmenterValue{
									{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
									{Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
								},
							},
						},
					},
					AllowedValues: &_segmenters.ListSegmenterValue{
						Values: []*_segmenters.SegmenterValue{
							{Value: &_segmenters.SegmenterValue_Real{Real: 1.5}},
							{Value: &_segmenters.SegmenterValue_Real{Real: 2.5}},
						},
					},
				},
			},
		},
	}
}

func TestSegmenter(t *testing.T) {
	suite.Run(t, new(SegmenterTestSuite))
}

func (s *SegmenterTestSuite) TestSegmenterGet() {
	t := s.Suite.T()
	seg1 := NewBaseSegmenter(s.SegmenterConfigs[0])
	seg2 := NewBaseSegmenter(s.SegmenterConfigs[1])
	// Validate
	assert.Equal(t, "test-segmenter-1", seg1.GetName())
	assert.Equal(t, _segmenters.SegmenterValueType_INTEGER, seg1.GetType())
	assert.Equal(t, _segmenters.SegmenterValueType_STRING, seg2.GetType())
	// Although TreatmentRequestFields is not set, GetExperimentVariables should return the segment name as default
	assert.Equal(t, &_segmenters.ListExperimentVariables{
		Values: []*_segmenters.ExperimentVariables{
			{
				Value: []string{"test-segmenter-1"},
			},
		},
	}, seg1.GetExperimentVariables())
	assert.Equal(t, &_segmenters.ListExperimentVariables{
		Values: []*_segmenters.ExperimentVariables{
			{
				Value: []string{"exp-var-1", "exp-var-2"},
			},
		},
	}, seg2.GetExperimentVariables())
	config, _ := seg1.GetConfiguration()
	assert.Equal(t, s.SegmenterConfigs[0], config)
}

func (s *SegmenterTestSuite) TestSegmenterIsValidType() {
	t := s.Suite.T()
	seg := NewBaseSegmenter(s.SegmenterConfigs[0])
	tests := map[string]struct {
		values  []*_segmenters.SegmenterValue
		success bool
	}{
		"failure | mixed types": {
			values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 10}},
				{Value: &_segmenters.SegmenterValue_Real{Real: 0.5}},
			},
		},
		"failure | invalid type": {
			values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Real{Real: 0.5}},
			},
		},
		"failure | empty list": {
			values:  []*_segmenters.SegmenterValue{},
			success: true,
		},
		"success | valid values": {
			values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 100}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 10}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 0}},
			},
			success: true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			success := seg.IsValidType(data.values)
			assert.Equal(t, data.success, success)
		})
	}
}

func (s *SegmenterTestSuite) TestSegmenterValidateSegmenterAndConstraints() {
	t := s.Suite.T()
	seg1 := NewBaseSegmenter(s.SegmenterConfigs[0])
	seg2 := NewBaseSegmenter(s.SegmenterConfigs[1])
	seg3 := NewBaseSegmenter(s.SegmenterConfigs[2])
	seg4 := NewBaseSegmenter(s.SegmenterConfigs[3])
	tests := map[string]struct {
		segmenter Segmenter
		values    map[string]*_segmenters.ListSegmenterValue
		errString string
	}{
		"success | empty map": {
			segmenter: seg1,
			values:    map[string]*_segmenters.ListSegmenterValue{},
		},
		"success | empty list": {
			segmenter: seg1,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-1": {
					Values: []*_segmenters.SegmenterValue{},
				},
			},
		},
		"failure | multi valued": {
			segmenter: seg2,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-2": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_String_{String_: "1"}},
						{Value: &_segmenters.SegmenterValue_String_{String_: "2"}},
					},
				},
			},
			errString: "Segmenter test-segmenter-2 is configured as single-valued but has multiple input values",
		},
		"failure | invalid value type": {
			segmenter: seg1,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-1": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_String_{String_: "2"}},
					},
				},
			},
			errString: "Segmenter test-segmenter-1 has one or more values that do not match the configured type",
		},
		"success | no options": {
			segmenter: seg3,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-3": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Real{Real: 1.5}},
					},
				},
			},
		},
		"success | no matching constraints": {
			segmenter: seg4,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-1": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 0}},
					},
				},
				"test-segmenter-4": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Real{Real: 10.5}},
					},
				},
			},
		},
		"failure | constaint check": {
			segmenter: seg4,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-1": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
					},
				},
				"test-segmenter-4": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Real{Real: 10.5}},
					},
				},
			},
			errString: "Values for segmenter test-segmenter-4 do not satisfy the constraint",
		},
		"success": {
			segmenter: seg4,
			values: map[string]*_segmenters.ListSegmenterValue{
				"test-segmenter-1": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
					},
				},
				"test-segmenter-4": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Real{Real: 1.5}},
					},
				},
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := data.segmenter.ValidateSegmenterAndConstraints(data.values)
			if data.errString == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}
