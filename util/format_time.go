package util

import "time"

// All time output are formatted in ISO8601 string,
// with timezone set in UTC.
// Example 2006-01-02T15:04:05Z
const (
	secondsOfMinute   = 60
	secondsOfHour     = 60 * secondsOfMinute
	layoutISO8601Date = "2006-01-02"
	layoutISO9075     = "2006-01-02 15:04:05" // Layout for SQL DATETIME
	layoutISO9075Date = "2006-01-02"          // Layout for SQL DATE
	layoutWxTime      = "20060102150405"
	layoutCST         = "2006年01月02日 15:04:05 中国标准时间"
	stmtUTC8Now       = "DATE_ADD(UTC_TIMESTAMP(), INTERVAL 8 HOUR)"
)

var (
	// TZShanghai is a fixed timezone in UTC+8
	TZShanghai = time.FixedZone("UTC+8", 8*secondsOfHour)
	// ToISO8601UTC turns time into ISO 8601 string in UTC.
	ToISO8601UTC = timeFormatter{time.RFC3339, time.UTC}
	// ToSQLDatetimeUTC turns time into SQL's DATETIME string in UTC.
	ToSQLDatetimeUTC = timeFormatter{layoutISO9075, time.UTC}
	// ToSQLDateUTC turns time into SQL's DATE string in UTC.
	ToSQLDateUTC = timeFormatter{layoutISO9075Date, time.UTC}
	// ToSQLDateUTC8 turns time into SQL's DATE string set in UTC+8.
	ToSQLDateUTC8 = timeFormatter{layoutISO9075Date, TZShanghai}
	// ToCST turns time into Chinese text set in Asia/Shanghai
	ToCST = timeFormatter{layoutCST, TZShanghai}
)

// timeFormatter converts a time.Time instance to the specified layout in specified location
type timeFormatter struct {
	layout string         // output layout
	loc    *time.Location // target timezone
}

// ToLocation changes a timeFormatter instance's
func (f timeFormatter) ToLocation(loc *time.Location) timeFormatter {
	f.loc = loc
	return f
}

// FromUnix formats a Unix timestamp to human readable string
func (f timeFormatter) FromUnix(sec int64) string {
	return time.Unix(sec, 0).In(f.loc).Format(f.layout)
}

// FromISO8601 parses a ISO8601 time string and returns the
// specified format, or returns the original string if parsing failed.
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

	t, err := time.ParseInLocation(layoutISO9075, value, loc)

	if err != nil {
		return "", err
	}

	return t.In(f.loc).Format(f.layout), nil
}

func (f timeFormatter) FromTime(t time.Time) string {
	return t.In(f.loc).Format(f.layout)
}

func (f timeFormatter) FromWx(value string) string {
	t, err := ParseWxTime(value)

	if err != nil {
		return ""
	}

	return t.In(f.loc).Format(f.layout)
}

// ParseSQLDate parse string layout `2006-01-02`
func ParseSQLDate(value string) (time.Time, error) {
	return time.Parse(layoutISO9075Date, value)
}

// ParseSQLDatetime parse SQL DATETIME string in UTC.
func ParseSQLDatetime(value string) time.Time {
	t, err := time.ParseInLocation(layoutISO9075, value, time.UTC)

	if err != nil {
		return time.Now()
	}

	return t
}

// ParseISO8601 parses ISO 8601 time string
// into a time.Time instance, or returns error.
func ParseISO8601(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// ParseWxTime is used to parse wxpay's time format.
// If it cannot be parsed, default to current time.
func ParseWxTime(value string) (time.Time, error) {
	return time.ParseInLocation(layoutWxTime, value, TZShanghai)
}

// ParseAliTime parses alipay time string.
// Not clear what timezone it uses. Assming Shanghai time.
func ParseAliTime(value string) (time.Time, error) {
	return time.ParseInLocation(layoutISO9075, value, TZShanghai)
}
