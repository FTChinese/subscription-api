package util

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ISODateTime is used to convert between ISO8601 format and SQL DATETIME format.
// The conversion is assumed to happen in UTC, e.g. the zero time zone.
// If your datetime is saved in another time zone,
// convert it manually.
type ISODateTime string

// UnmarshalJSON implements the Unmarshaler interface to parse JSON into ISODateTime type.
func (d *ISODateTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*d = ISODateTime(s)

	return nil
}

// MarshalJSON implements Marshaler interface to stringify a value.
func (d ISODateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

// Scan converts SQL Datetime to ISO8601 format.
// If input is a SQL null, target will have a zero value, which is an empty string.
func (d *ISODateTime) Scan(src interface{}) error {
	// Handling SQL null value.
	if src == nil {
		*d = ISODateTime("")
		return nil
	}

	switch s := src.(type) {
	case []byte:
		dt, err := ToISO8601UTC.FromDatetime(string(s), nil)

		if err != nil {
			return err
		}

		*d = ISODateTime(dt)
		return nil

	default:
		return ErrIncompatible
	}
}

// Value converts ISO8601 to SQL Datetime
func (d ISODateTime) Value() (driver.Value, error) {
	s := string(d)

	if s == "" {
		return nil, nil
	}

	dt, err := ToSQLDatetimeUTC.FromISO8601(s)

	// If input string cannot be parsed as ISO8601 format.
	if err != nil {
		return nil, err
	}

	return dt, nil
}

// ToTime parses an instance.
func (d ISODateTime) ToTime() (time.Time, error) {
	return ParseISO8601(string(d))
}
