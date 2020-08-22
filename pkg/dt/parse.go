package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"time"
)

const (
	// Wechat's time layout format.
	layoutWx = "20060102150405"
)

// ParseWxTime is used to parse wxpay's time format.
// If it cannot be parsed, default to current time.
func ParseWxTime(value string) (time.Time, error) {
	t, err := time.ParseInLocation(layoutWx, value, chrono.TZShanghai)
	if err != nil {
		return t, err
	}

	return t, nil
}

// MustParseWxTime parses wechat pay time format,
// and falls back to now if parse failed.
func MustParseWxTime(value string) time.Time {
	t, err := ParseWxTime(value)
	if err != nil {
		return time.Now()
	}

	return t
}

// ParseAliTime parses alipay time string.
// Not clear what timezone it uses. Assming Shanghai time.
func ParseAliTime(value string) (time.Time, error) {
	t, err := time.ParseInLocation(chrono.SQLDateTime, value, chrono.TZShanghai)
	if err != nil {
		return t, err
	}

	return t, nil
}

// MustParseAliTime parses alipay time format,
// and falls back to now if parse failed.
func MustParseAliTime(value string) time.Time {
	t, err := ParseAliTime(value)
	if err != nil {
		return time.Now()
	}

	return t
}
