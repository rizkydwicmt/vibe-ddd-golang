package timehelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMapDateTimeValue(t *testing.T) {
	tm := time.Now().UTC().Truncate(time.Second)
	header := map[string]interface{}{
		"date": tm.Format(time.RFC3339),
	}
	result := GetMapDateTimeValue(header, "date")
	assert.NotNil(t, result)
	assert.Equal(t, tm, *result)

	nilResult := GetMapDateTimeValue(header, "notfound")
	assert.Nil(t, nilResult)

	header["invalid"] = "not-a-date"
	invalidResult := GetMapDateTimeValue(header, "invalid")
	assert.Nil(t, invalidResult)
}

func TestParseDateTime(t *testing.T) {
	dt := "2024-06-20T15:04:05Z"
	expected, _ := time.Parse(time.RFC3339, dt)
	parsed, err := ParseDateTime(dt)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	_, err = ParseDateTime("salah-format")
	assert.Error(t, err)
}

func TestToUnixTimestamp(t *testing.T) {
	tm := time.Date(2024, 6, 20, 15, 4, 5, 0, time.UTC)
	ts := ToUnixTimestamp(tm)
	assert.Equal(t, tm.Unix(), ts)
}

func TestFormatDateWithTime(t *testing.T) {
	// FormatDateWithTime renders in the process-local zone; pin it so the
	// expectation holds on any machine (CI runs UTC).
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)
	orig := time.Local
	time.Local = loc
	defer func() { time.Local = orig }()

	ts := int64(1750666336)
	formatted := FormatDateWithTime(ts)
	assert.Equal(t, "2025-06-23 15:12:16", formatted)
}
