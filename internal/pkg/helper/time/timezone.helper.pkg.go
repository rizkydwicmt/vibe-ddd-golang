package timehelper

import (
	"time"
)

func GetTimezone(timezone string) string {
	if timezone != "" {
		return timezone
	}
	return "Asia/Jakarta"
}

func GetDateStart(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(GetTimezone(timezone))
	if err != nil {
		return time.Time{}, err
	}

	timeInTZ := t.In(loc)
	dateStart := time.Date(
		timeInTZ.Year(),
		timeInTZ.Month(),
		timeInTZ.Day(),
		0, 0, 0, 0,
		loc,
	)

	return dateStart, nil
}
