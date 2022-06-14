package segmenters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/xp/common/segmenters"
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

func (s *RunnersTestSuite) TestRunnerGet() {
	t := s.Suite.T()
	runner := NewBaseRunner(s.segmenterConfigs[0])

	assert.Equal(t, "test-runner", runner.GetName())
}

func (s *RunnersTestSuite) TestRunnerTransform() {
	t := s.Suite.T()
	runner := NewBaseRunner(s.segmenterConfigs[0])
	expected := []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}}}

	res, err := runner.Transform("test-seg", map[string]interface{}{"test-seg": 1}, []string{"test-seg"})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}
