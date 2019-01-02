package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Date is used to save and output YYYY-MM-DD format date string.
type Date struct {
	time.Time
}

// MarshalJSON converts a Time struct to ISO8601 string.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(ToSQLDateUTC.FromTime(d.Time))
}

// UnmarshalJSON converts ISO8601 data time into a Time struct.
// Empty string and null will be turned into time.Time zero value.
func (d *Date) UnmarshalJSON(data []byte) (err error) {

	fmt.Printf("Unmarshal data: %s\n", data)

	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}

	switch x := v.(type) {
	case string:
		if strings.TrimSpace(x) == "" {
			d.Time = time.Time{}
			return
		}

		d.Time, err = time.Parse(layoutDate, x)

		return
	case nil:
		d.Time = time.Time{}
		return
	}

	return fmt.Errorf("json: canot unmarshal %v into Go value of type Time", reflect.TypeOf(v).Name())
}

// Scan implements the Scanner interface.
// SQL NULL will be turned into time zero value.
func (d *Date) Scan(value interface{}) (err error) {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return
	case []byte:
		d.Time, err = time.Parse(layoutDate, string(v))
		return
	case string:
		d.Time, err = time.Parse(layoutDate, v)
		return
	}

	return fmt.Errorf("Can't convert %T to time.Time", value)
}

// Value implements the driver Valuer interface.
// Zero value is turned into SQL NULL.
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}

	return ToSQLDateUTC.FromTime(d.Time), nil
}

// DateNow creates current time.
func DateNow() Date {
	return Date{
		time.Now(),
	}
}

// DateZero creates the zero value of Time.
func DateZero() Date {
	return Date{time.Time{}}
}

// DateFrom creates a new Time wrapping time.Time.
func DateFrom(t time.Time) Date {
	return Date{t}
}
