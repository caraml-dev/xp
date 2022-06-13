package segmenters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/treatment-service/util"
)

type TimeRunnerTestSuite struct {
	suite.Suite

	hoursOfDayRunner Runner
	daysOfWeekRunner Runner
	name             string
	config           map[string]interface{}
}

func (suite *TimeRunnerTestSuite) SetupSuite() {
	suite.config = map[string]interface{}{}
	suite.name = "time"

	configJSON, err := json.Marshal(suite.config)
	suite.Require().NoError(err)
	s, err := NewHoursOfDaySegmenter(configJSON)
	suite.Require().NoError(err)
	suite.hoursOfDayRunner = s

	s, err = NewDaysOfWeekSegmenter(configJSON)
	suite.Require().NoError(err)
	suite.daysOfWeekRunner = s
}

func TestTimeRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(TimeRunnerTestSuite))
}

func (s *TimeRunnerTestSuite) TestHourOfDayTransform() {
	t := s.Suite.T()

	tzString := "Asia/Singapore"
	timeLoc, _ := util.RetrieveTimezone(tzString)
	hourOfDay := util.RetrieveHourOfDay(*timeLoc)

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
				"tz": tzString,
			},
			experimentVariables: []string{"invalid_var"},
			errString:           fmt.Sprintf("no valid variables were provided for %s segmenter", s.name),
		},
		{
			name: "failure | invalid hour_of_day",
			requestParam: map[string]interface{}{
				"hour_of_day": float64(24),
			},
			experimentVariables: []string{"hour_of_day"},
			errString:           fmt.Sprintf("provided hour_of_day variable for %s segmenter is invalid", s.name),
		},
		{
			name: "failure | invalid type tz variable",
			requestParam: map[string]interface{}{
				"tz": int64(20),
			},
			experimentVariables: []string{"tz"},
			errString:           fmt.Sprintf(TypeCastingErrorTmpl, "tz", s.name, "string"),
		},
		{
			name: "failure | invalid type hour_of_day variable",
			requestParam: map[string]interface{}{
				"hour_of_day": "1",
			},
			experimentVariables: []string{"hour_of_day"},
			errString:           fmt.Sprintf(TypeCastingErrorTmpl, "hour_of_day", s.name, "float64"),
		},
		{
			name: "success | tz",
			requestParam: map[string]interface{}{
				"tz": tzString,
			},
			experimentVariables: []string{"tz"},
			expected:            []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: hourOfDay}}},
		},
		{
			name: "success | hour_of_day",
			requestParam: map[string]interface{}{
				"hour_of_day": float64(hourOfDay),
			},
			experimentVariables: []string{"hour_of_day"},
			expected:            []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: hourOfDay}}},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			transformation, err := s.hoursOfDayRunner.Transform(s.name, data.requestParam, data.experimentVariables)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Require().Equal(data.expected, transformation)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *TimeRunnerTestSuite) TestDayOfWeekTransform() {
	t := s.Suite.T()

	tzString := "Asia/Singapore"
	timeLoc, _ := util.RetrieveTimezone(tzString)
	dayOfWeek := util.RetrieveDayOfWeek(*timeLoc)

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
				"tz": tzString,
			},
			experimentVariables: []string{"invalid_var"},
			errString:           fmt.Sprintf("no valid variables were provided for %s segmenter", s.name),
		},
		{
			name: "failure | invalid day_of_week",
			requestParam: map[string]interface{}{
				"day_of_week": float64(0),
			},
			experimentVariables: []string{"day_of_week"},
			errString:           fmt.Sprintf("provided day_of_week variable for %s segmenter is invalid", s.name),
		},
		{
			name: "failure | invalid type tz variable",
			requestParam: map[string]interface{}{
				"tz": int64(20),
			},
			experimentVariables: []string{"tz"},
			errString:           fmt.Sprintf(TypeCastingErrorTmpl, "tz", s.name, "string"),
		},
		{
			name: "failure | invalid type day_of_week variable",
			requestParam: map[string]interface{}{
				"day_of_week": "1",
			},
			experimentVariables: []string{"day_of_week"},
			errString:           fmt.Sprintf(TypeCastingErrorTmpl, "day_of_week", s.name, "float64"),
		},
		{
			name: "success | tz",
			requestParam: map[string]interface{}{
				"tz": tzString,
			},
			experimentVariables: []string{"tz"},
			expected:            []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: dayOfWeek}}},
		},
		{
			name: "success | day_of_week",
			requestParam: map[string]interface{}{
				"day_of_week": float64(dayOfWeek),
			},
			experimentVariables: []string{"day_of_week"},
			expected:            []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: dayOfWeek}}},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			transformation, err := s.daysOfWeekRunner.Transform(s.name, data.requestParam, data.experimentVariables)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Require().Equal(data.expected, transformation)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}
