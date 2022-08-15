package segmenters

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

func NewHoursOfDaySegmenter(_ json.RawMessage) (Segmenter, error) {
	var hoursOfDayConfig = &_segmenters.SegmenterConfiguration{
		Name: "hours_of_day",
		Type: _segmenters.SegmenterValueType_INTEGER,
		Options: func() map[string]*_segmenters.SegmenterValue {
			options := map[string]*_segmenters.SegmenterValue{}
			for i := 0; i < 24; i++ {
				options[fmt.Sprint(i)] = &_segmenters.SegmenterValue{
					Value: &_segmenters.SegmenterValue_Integer{
						Integer: int64(i),
					},
				}
			}
			return options
		}(),
		MultiValued: true,
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{"hour_of_day"},
				},
				{
					Value: []string{"tz"},
				},
			},
		},
		Required: false,
	}

	return NewBaseSegmenter(hoursOfDayConfig), nil
}

func NewDaysOfWeekSegmenter(_ json.RawMessage) (Segmenter, error) {
	var daysOfWeekConfig = &_segmenters.SegmenterConfiguration{
		Name: "days_of_week",
		Type: _segmenters.SegmenterValueType_INTEGER,
		Options: func() map[string]*_segmenters.SegmenterValue {
			options := map[string]*_segmenters.SegmenterValue{}
			for i := 1; i <= 7; i++ {
				// TODO: The Clojure API uses Monday = 1 to Sunday = 7, whereas Golang's
				// time library uses Sunday = 0, Monday = 1, etc. So, the days are being
				// computed using i=1 to 7 and doing %7. This can be removed in the future,
				// when the Clojure APIs have been deprecated.
				weekday := time.Weekday(i % 7)
				options[weekday.String()] = &_segmenters.SegmenterValue{
					Value: &_segmenters.SegmenterValue_Integer{
						Integer: int64(i),
					},
				}
			}
			return options
		}(),
		MultiValued: true,
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{"day_of_week"},
				},
				{
					Value: []string{"tz"},
				},
			},
		},
		Required: false,
	}

	return NewBaseSegmenter(daysOfWeekConfig), nil
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
