package util

import (
	"database/sql/driver"
	"errors"
)

// Datetime is used to convert between ISO8601 and SQL DATETIME format.
type Datetime string

// Scan converts SQL Datetime to ISO8601 format
func (d *Datetime) Scan(src interface{}) error {
	var source string
	switch src.(type) {
	case string:
		source = src.(string)
	default:
		return errors.New("imcompatible type for datetime")
	}

	if source == "" {
		*d = ""
		return nil
	}

	dt := ISO8601UTC.FromDatetime(source, nil)

	*d = Datetime(dt)
	return nil
}

// Value converts ISO8601 to SQL Datetime
func (d Datetime) Value() (driver.Value, error) {
	s := string(d)

	if s == "" {
		return driver.Value(""), nil
	}

	dt := SQLDatetimeUTC.FromISO8601(s)

	return driver.Value(dt), nil
}
