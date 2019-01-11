package util

import (
	"fmt"
	"time"
)

// All time output are formatted in ISO8601 string,
// with timezone set in UTC.
// Example 2006-01-02T15:04:05Z
const (
	secondsOfMinute = 60
	secondsOfHour   = 60 * secondsOfMinute
	layoutDateTime  = "2006-01-02 15:04:05.999999"
	layoutWx        = "20060102150405"
)

var (
	// TZShanghai is a fixed timezone in UTC+8
	TZShanghai = time.FixedZone("UTC+8", 8*secondsOfHour)
	// ToISO8601UTC turns time into ISO 8601 string in UTC.
	ToISO8601UTC = timeFormatter{time.RFC3339, time.UTC}
	// ToSQLDatetimeUTC turns time into SQL's DATETIME string in UTC.
	ToSQLDatetimeUTC = timeFormatter{layoutDateTime[:19], time.UTC}
	// ToSQLDateUTC turns time into SQL's DATE string in UTC.
	ToSQLDateUTC = timeFormatter{layoutDateTime[:10], time.UTC}
	// ToSQLDateUTC8 turns time into SQL's DATE string set in UTC+8.
	ToSQLDateUTC8 = timeFormatter{layoutDateTime[:10], TZShanghai}
)

// timeFormatter converts a time.Time instance to the specified layout in specified location
type timeFormatter struct {
	layout string         // output layout
	loc    *time.Location // target timezone
}

// ToLocation changes a timeFormatter instance's.
// Returns a new instance of timeFormatter.
func (f timeFormatter) ToLocation(loc *time.Location) timeFormatter {
	f.loc = loc
	return f
}

// FromUnix formats a Unix timestamp to human readable string
func (f timeFormatter) FromUnix(sec int64) string {
	return time.Unix(sec, 0).In(f.loc).Format(f.layout)
}

// FromISO8601 parses a ISO8601 time string and returns the
// specified format.
func (f timeFormatter) FromISO8601(value string) (string, error) {
	t, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return "", err
	}

	return t.In(f.loc).Format(f.layout), nil
}

// FromDatetime formats SQL DATETIME.
// Parameter `loc` is input string's location since SQL DATETIME do not have time zone information.
// If loc is nil, defaults to UTC.
func (f timeFormatter) FromDatetime(value string, loc *time.Location) (string, error) {
	if loc == nil {
		loc = time.UTC
	}

	t, err := time.ParseInLocation(layoutDateTime[:len(value)], value, loc)

	if err != nil {
		return "", err
	}

	return t.In(f.loc).Format(f.layout), nil
}

func (f timeFormatter) FromTime(t time.Time) string {
	return t.In(f.loc).Format(f.layout)
}

// ParseDateTime parses SQL DATE or DATETIME string in specified location.
func ParseDateTime(str string, loc *time.Location) (t time.Time, err error) {
	base := "0000-00-00 00:00:00.0000000"
	switch len(str) {
	case 10, 19: // up to "YYYY-MM-DD HH:MM:SS"
		if str == base[:len(str)] {
			return
		}
		t, err = time.Parse(layoutDateTime[:len(str)], str)
	default:
		err = fmt.Errorf("invalid time string: %s", str)
		return
	}

	// Adjust location
	if err == nil && loc != time.UTC {
		y, mo, d := t.Date()
		h, mi, s := t.Clock()
		t, err = time.Date(y, mo, d, h, mi, s, t.Nanosecond(), loc), nil
	}

	return
}

// ParseWxTime is used to parse wxpay's time format.
// If it cannot be parsed, default to current time.
func ParseWxTime(value string) time.Time {
	t, err := time.ParseInLocation(layoutWx, value, TZShanghai)
	if err != nil {
		return time.Now()
	}

	return t
}

// ParseAliTime parses alipay time string.
// Not clear what timezone it uses. Assming Shanghai time.
func ParseAliTime(value string) (time.Time, error) {
	return ParseDateTime(value, TZShanghai)
}
