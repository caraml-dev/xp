package segmenters

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cast"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/util"
)

var (
	hourMin int64 = 0
	hourMax int64 = 23
	dayMin  int64 = 1
	dayMax  int64 = 7
)

func NewHoursOfDaySegmenter(_ json.RawMessage) (Runner, error) {
	var hoursOfDayConfig = &SegmenterConfig{
		Name: "hours_of_day",
	}

	return &hoursOfDay{NewBaseRunner(hoursOfDayConfig)}, nil
}

type hoursOfDay struct {
	Runner
}

func (s *hoursOfDay) Transform(
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	var hourOfDay int64
	var err error
	switch {
	case cmp.Equal(experimentVariables, []string{"tz"}):
		tzString, ok := requestValues["tz"].(string)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, "tz", segmenter, "string")
		}
		timeLoc, err := util.RetrieveTimezone(tzString)
		if err != nil {
			return nil, err
		}
		hourOfDay = util.RetrieveHourOfDay(*timeLoc)
	case cmp.Equal(experimentVariables, []string{"hour_of_day"}):
		hourOfDay, err = cast.ToInt64E(requestValues["hour_of_day"])
		if err != nil {
			return nil, err
		}
		if hourOfDay < hourMin || hourOfDay > hourMax {
			return nil, fmt.Errorf("provided hour_of_day variable for %s segmenter is invalid", segmenter)
		}
	default:
		return nil, fmt.Errorf("no valid variables were provided for %s segmenter", segmenter)
	}
	segmenterValue := []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: hourOfDay}}}

	return segmenterValue, nil
}

func NewDaysOfWeekSegmenter(_ json.RawMessage) (Runner, error) {
	var daysOfWeekConfig = &SegmenterConfig{
		Name: "days_of_week",
	}

	return &daysOfWeek{NewBaseRunner(daysOfWeekConfig)}, nil
}

type daysOfWeek struct {
	Runner
}

func (s *daysOfWeek) Transform(
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	var dayOfWeek int64
	var err error
	switch {
	case cmp.Equal(experimentVariables, []string{"tz"}):
		tzString, ok := requestValues["tz"].(string)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, "tz", segmenter, "string")
		}
		timeLoc, err := util.RetrieveTimezone(tzString)
		if err != nil {
			return nil, err
		}
		dayOfWeek = util.RetrieveDayOfWeek(*timeLoc)
	case cmp.Equal(experimentVariables, []string{"day_of_week"}):
		dayOfWeek, err = cast.ToInt64E(requestValues["day_of_week"])
		if err != nil {
			return nil, err
		}
		if dayOfWeek < dayMin || dayOfWeek > dayMax {
			return nil, fmt.Errorf("provided day_of_week variable for %s segmenter is invalid", segmenter)
		}
	default:
		return nil, fmt.Errorf("no valid variables were provided for %s segmenter", segmenter)
	}
	segmenterValue := []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: dayOfWeek}}}

	return segmenterValue, nil
}

func init() {
	err := Register("hours_of_day", NewHoursOfDaySegmenter)
	if err != nil {
		log.Fatal(err)
	}
	err = Register("days_of_week", NewDaysOfWeekSegmenter)
	if err != nil {
		log.Fatal(err)
	}
}
