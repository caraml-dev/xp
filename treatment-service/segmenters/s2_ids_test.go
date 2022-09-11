package segmenters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type S2IDsRunnerTestSuite struct {
	suite.Suite

	runner Runner
	name   string
	config map[string]interface{}
}

func (suite *S2IDsRunnerTestSuite) SetupSuite() {
	suite.config = map[string]interface{}{
		"mins2celllevel": 10,
		"maxs2celllevel": 14,
	}
	suite.name = "s2_ids"

	configJSON, err := json.Marshal(suite.config)
	suite.Require().NoError(err)
	s, err := NewS2IDRunner(configJSON)
	suite.Require().NoError(err)
	suite.runner = s
}

func TestS2IDsRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(S2IDsRunnerTestSuite))
}

func (s *S2IDsRunnerTestSuite) TestTransform() {
	t := s.Suite.T()

	tests := []struct {
		name                string
		requestParam        map[string]interface{}
		experimentVariables []string
		expected            []*_segmenters.SegmenterValue
		errString           string
	}{
		{
			name: "failure | no valid variable",
			requestParam: map[string]interface{}{
				"s2id": 3348536261227839488,
			},
			experimentVariables: []string{"invalid_var"},
			errString:           fmt.Sprintf("no valid variables were provided for %s segmenter", s.name),
		},
		{
			name: "failure | invalid s2id",
			requestParam: map[string]interface{}{
				"s2id": float64(3348536),
			},
			experimentVariables: []string{"s2id"},
			errString:           fmt.Sprintf("provided s2id variable for %s segmenter is invalid", s.name),
		},
		{
			name: "failure | invalid type s2id variable",
			requestParam: map[string]interface{}{
				"s2id": int64(3348536),
			},
			experimentVariables: []string{"s2id"},
			errString:           fmt.Sprintf("provided s2id variable for %s segmenter is invalid", s.name),
		},
		{
			name: "failure | invalid type latitude variable",
			requestParam: map[string]interface{}{
				"latitude":  int64(106),
				"longitude": 103.899899113748,
			},
			experimentVariables: []string{"latitude", "longitude"},
			errString:           "received invalid latitude, longitude values",
		},
		{
			name: "success | lat-long + ordering",
			requestParam: map[string]interface{}{
				"latitude":  1.2537040223936706,
				"longitude": 103.899899113748,
			},
			experimentVariables: []string{"longitude", "latitude"},
			expected: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210809859604480}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210814154571776}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210796974702592}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210865694179328}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592211140572086272}},
			},
		},
		{
			name: "success | lat-long + ordering (string)",
			requestParam: map[string]interface{}{
				"latitude":  "1.2537040223936706",
				"longitude": "103.899899113748",
			},
			experimentVariables: []string{"longitude", "latitude"},
			expected: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210809859604480}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210814154571776}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210796974702592}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592210865694179328}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592211140572086272}},
			},
		},
		{
			name: "success | s2id",
			requestParam: map[string]interface{}{
				"s2id": float64(3348536261227839488),
			},
			experimentVariables: []string{"s2id"},
			expected: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3348536261227839488}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3348536256932872192}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3348536205393264640}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3348535999234834432}},
				{Value: &_segmenters.SegmenterValue_Integer{Integer: 3348535174601113600}},
			},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			transformation, err := s.runner.Transform(s.name, data.requestParam, data.experimentVariables)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Require().Equal(data.expected, transformation)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}
