package timehelper

import (
	"fmt"
	"time"
)

func GetMapDateTimeValue(header map[string]interface{}, key string) *time.Time {
	value, exists := header[key]
	if !exists || value == nil {
		return nil
	}
	strValue := fmt.Sprintf("%v", value)
	parsedTime, err := time.Parse(time.RFC3339, strValue)
	if err != nil {
		return nil
	}
	return &parsedTime
}

func ParseDateTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}

func ToUnixTimestamp(date time.Time) int64 {
	return date.Unix()
}

func FormatDateWithTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}
