package util

import "time"

const (
	secondsOfMinute = 60
	secondsOfHour   = 60 * secondsOfMinute
	iso8601Date     = "2006-01-02"
	iso9075         = "2006-01-02 15:04:05" // Layout for SQL DATETIME
	iso9075Date     = "2006-01-02"          // Layout for SQL DATE
	layoutWxTime    = "20060102150405"
	stmtUTC8Now     = "DATE_ADD(UTC_TIMESTAMP(), INTERVAL 8 HOUR)"
)

var (
	// TZShanghai is a fixed timezone in UTC+8
	TZShanghai = time.FixedZone("UTC+8", 8*secondsOfHour)
	// ISO8601UTC turns time into ISO 8601 string in UTC.
	ISO8601UTC = timeForamtter{time.RFC3339, time.UTC}
	// SQLDatetimeUTC turns time into SQL's DATETIME string in UTC.
	SQLDatetimeUTC = timeForamtter{iso9075, time.UTC}
	// SQLDateUTC turns time into SQL's DATE string in UTC.
	SQLDateUTC = timeForamtter{iso9075Date, time.UTC}
	// SQLDateUTC8 turns time into SQL's DATE string set in UTC+8.
	SQLDateUTC8 = timeForamtter{iso9075Date, TZShanghai}
)

// timeFormatter converts a time.Time instance to the specified layout in specified location
type timeForamtter struct {
	layout string         // output layout
	loc    *time.Location // destination timezone
}

func (f timeForamtter) ToLocation(loc *time.Location) timeForamtter {
	f.loc = loc
	return f
}

// FromUnix formats a Unix timestamp to human readable string
func (f timeForamtter) FromUnix(sec int64) string {
	return time.Unix(sec, 0).In(f.loc).Format(f.layout)
}

// FromDatetime formats SQL DATETIME.
// Parameter `loc` is input string's location since SQL DATETIME do not have time zone information.
// If loc is nil, defaults to UTC.
func (f timeForamtter) FromDatetime(value string, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}

	t, err := time.ParseInLocation(iso9075, value, loc)

	if err != nil {
		return value
	}

	return t.In(f.loc).Format(f.layout)
}

func (f timeForamtter) FromTime(t time.Time) string {
	return t.In(f.loc).Format(f.layout)
}

func (f timeForamtter) FromWx(value string) string {
	t, err := ParseWxTime(value)

	if err != nil {
		return ""
	}

	return t.In(f.loc).Format(f.layout)
}

// ParseSQLDate parse string layout `2006-01-02`
func ParseSQLDate(value string) (time.Time, error) {
	return time.Parse(iso9075Date, value)
}

// ParseSQLDatetime parse SQL DATETIME string in UTC.
func ParseSQLDatetime(value string) time.Time {
	t, err := time.ParseInLocation(iso9075, value, time.UTC)

	if err != nil {
		return time.Now()
	}

	return t
}

// ParseWxTime is used to parse wxpay's time format.
// If it cannot be parsed, default to current time.
func ParseWxTime(value string) (time.Time, error) {
	return time.ParseInLocation(layoutWxTime, value, TZShanghai)
}

// ParseAliTime parses alipay time string.
// Not clear what timezone it uses. Assming Shanghai time.
func ParseAliTime(value string) (time.Time, error) {
	return time.ParseInLocation(iso9075, value, TZShanghai)
}
