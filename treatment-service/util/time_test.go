package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetrieveTimezone(t *testing.T) {
	resp, err := RetrieveTimezone("Asia/Singapore")
	expected, _ := time.LoadLocation("Asia/Singapore")
	assert.Equal(t, expected, resp)
	assert.NoError(t, nil, err)

	_, err = RetrieveTimezone("Asia/dummy")
	assert.Error(t, err, "unknown time zone")
}

func TestRetrieveHourOfDay(t *testing.T) {
	tz, _ := time.LoadLocation("Asia/Singapore")
	resp := RetrieveHourOfDay(*tz)
	expected := int64(time.Now().In(tz).Hour())
	assert.Equal(t, expected, resp)
}

func TestRetrieveDayOfWeek(t *testing.T) {
	tz, _ := time.LoadLocation("Asia/Singapore")
	resp := RetrieveDayOfWeek(*tz)

	weekdayMap := map[string]int64{
		"Monday":    1,
		"Tuesday":   2,
		"Wednesday": 3,
		"Thursday":  4,
		"Friday":    5,
		"Saturday":  6,
		"Sunday":    7,
	}
	now := time.Now().In(tz)
	expected := weekdayMap[now.Weekday().String()]

	assert.Equal(t, expected, resp)
}

func TestWaitFor(t *testing.T) {
	start := time.Now()
	ok := WaitFor(func() bool {
		t1 := time.Since(start).Seconds()
		return t1 > 1
	}, time.Second*time.Duration(3), time.Second)
	assert.Equal(t, true, ok)

	start = time.Now()
	ok = WaitFor(func() bool {
		t1 := time.Since(start).Seconds()
		return t1 > 4
	}, time.Second*time.Duration(3), time.Second)
	assert.Equal(t, false, ok)
}
