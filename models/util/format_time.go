package util

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
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

// ParseAliTime parses alipay time string.
// Not clear what timezone it uses. Assming Shanghai time.
func ParseAliTime(value string) time.Time {
	t, err := time.ParseInLocation(chrono.SQLDateTime, value, chrono.TZShanghai)
	if err != nil {
		return time.Now()
	}

	return t
}
