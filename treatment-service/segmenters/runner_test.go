package segmenters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type RunnersTestSuite struct {
	suite.Suite
	segmenterConfigs []*SegmenterConfig
}

func (s *RunnersTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up RunnerTestSuite")

	s.segmenterConfigs = []*SegmenterConfig{
		{Name: "test-runner"},
	}
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnersTestSuite))
}

func (s *RunnersTestSuite) TestBaseRunnerGet() {
	t := s.Suite.T()
	runner := NewBaseRunner(s.segmenterConfigs[0])

	assert.Equal(t, "test-runner", runner.GetName())
}

func (s *RunnersTestSuite) TestBaseRunnerTransform() {
	t := s.Suite.T()
	segmenterName := "test-seg"
	protoStringType := _segmenters.SegmenterValueType_STRING
	protoBoolType := _segmenters.SegmenterValueType_BOOL
	protoIntegerType := _segmenters.SegmenterValueType_INTEGER
	protoRealType := _segmenters.SegmenterValueType_REAL
	tests := []struct {
		testName      string
		requestValues map[string]interface{}
		segmenterType *_segmenters.SegmenterValueType
		expected      []*_segmenters.SegmenterValue
		errString     string
	}{
		{
			testName:      "success | default untyped string inferred",
			requestValues: map[string]interface{}{segmenterName: "1"},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "1"}}},
		},
		{
			testName:      "success | default untyped, int inferred",
			requestValues: map[string]interface{}{segmenterName: 1},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}}},
		},
		{
			testName:      "success | default untyped float inferred",
			requestValues: map[string]interface{}{segmenterName: 1.1},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Real{Real: 1.1}}},
		},
		{
			testName:      "success | default untyped bool inferred",
			requestValues: map[string]interface{}{segmenterName: false},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: false}}},
		},
		{
			testName: "success | integer type",
			// using float as JSON value via browser are sent as float
			requestValues: map[string]interface{}{segmenterName: float64(1)},
			segmenterType: &protoIntegerType,
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}}},
		},
		{
			testName: "failure | integer type",
			// using float as JSON value via browser are sent as float
			requestValues: map[string]interface{}{segmenterName: "string_segmenter"},
			segmenterType: &protoIntegerType,
			errString:     "unable to cast \"string_segmenter\" of type string to int64",
		},
		{
			testName:      "success | string type",
			requestValues: map[string]interface{}{segmenterName: "1"},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "1"}}},
			segmenterType: &protoStringType,
		},
		{
			testName:      "failure | string type",
			requestValues: map[string]interface{}{segmenterName: 1},
			segmenterType: &protoStringType,
			errString:     "segmenter type for test-seg is not supported",
		},
		{
			testName:      "success | float type",
			requestValues: map[string]interface{}{segmenterName: 1.1},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Real{Real: 1.1}}},
			segmenterType: &protoRealType,
		},
		{
			testName:      "failure | float type",
			requestValues: map[string]interface{}{segmenterName: "string_segmenter"},
			segmenterType: &protoRealType,
			errString:     "unable to cast \"string_segmenter\" of type string to float64",
		},
		{
			testName:      "success | bool type",
			requestValues: map[string]interface{}{segmenterName: false},
			expected:      []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: false}}},
			segmenterType: &protoBoolType,
		},
		{
			testName:      "failure | bool type",
			requestValues: map[string]interface{}{segmenterName: "string"},
			segmenterType: &protoBoolType,
			errString:     "strconv.ParseBool: parsing \"string\": invalid syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			runner := NewBaseRunner(&SegmenterConfig{
				Name: segmenterName,
				Type: test.segmenterType,
			})
			res, err := runner.Transform(segmenterName, test.requestValues, []string{segmenterName})
			if test.errString == "" {
				s.Assert().NoError(err)
				s.Assert().Equal(test.expected, res)
			} else {
				s.Assert().EqualError(err, test.errString)
			}
		})
	}
}
