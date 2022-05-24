package util

import "time"

func RetrieveTimezone(tz interface{}) (*time.Location, error) {
	timezone, err := time.LoadLocation(tz.(string))
	if err != nil {
		return nil, err
	}
	return timezone, nil
}

func RetrieveHourOfDay(tz time.Location) int64 {
	now := time.Now().In(&tz)
	return int64(now.Hour())
}

func RetrieveDayOfWeek(tz time.Location) int64 {
	weekdayMap := map[string]int64{
		"Monday":    1,
		"Tuesday":   2,
		"Wednesday": 3,
		"Thursday":  4,
		"Friday":    5,
		"Saturday":  6,
		"Sunday":    7,
	}
	now := time.Now().In(&tz)
	dayOfWeek := weekdayMap[now.Weekday().String()]

	return dayOfWeek
}

func WaitFor(condition func() bool, waitFor time.Duration, tick time.Duration) bool {
	ch := make(chan bool, 1)

	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			return false
		case <-tick:
			tick = nil
			go func() { ch <- condition() }()
		case v := <-ch:
			if v {
				return true
			}
			tick = ticker.C
		}
	}
}
