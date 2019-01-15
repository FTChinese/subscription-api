package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const layoutCST = "2006年01月02日 15:04:05 中国标准时间"

// Time is used to save and output ISO8601 date time.
type Time struct {
	time.Time
}

// StringEN produces the string representation in English with locale set to UTC.
func (t Time) StringEN() string {
	return t.In(time.UTC).Format(time.RFC1123Z)
}

// StringCN produces the string representation in Chinese format with locale set to Asia/Shanghai.
func (t Time) StringCN() string {
	return t.In(TZShanghai).Format(layoutCST)
}

// MarshalJSON converts a Time struct to ISO8601 string.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t.In(time.UTC).Format(time.RFC3339))
}

// UnmarshalJSON converts ISO8601 data time into a Time struct.
// Empty string and null will be turned into time.Time zero value.
func (t *Time) UnmarshalJSON(data []byte) (err error) {

	fmt.Printf("Unmarshal data: %s\n", data)

	var v interface{}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}

	switch x := v.(type) {
	case string:
		if strings.TrimSpace(x) == "" {
			t.Time = time.Time{}
			return
		}

		t.Time, err = time.Parse(time.RFC3339, x)

		return
	case nil:
		t.Time = time.Time{}
		return
	}

	return fmt.Errorf("json: canot unmarshal %v into Go value of type Time", reflect.TypeOf(v).Name())
}

// Scan implements the Scanner interface.
// SQL NULL will be turned into time zero value.
func (t *Time) Scan(value interface{}) (err error) {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return
	case []byte:
		t.Time, err = ParseDateTime(string(v), time.UTC)
		return
	case string:
		t.Time, err = ParseDateTime(v, time.UTC)
		return
	}

	return fmt.Errorf("Can't convert %T to time.Time", value)
}

// Value implements the driver Valuer interface.
// Zero value is turned into SQL NULL.
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}

	return t.In(time.UTC).Format(layoutDateTime), nil
}

// TimeNow creates current time.
func TimeNow() Time {
	return Time{
		time.Now(),
	}
}

// TimeZero creates the zero value of Time.
func TimeZero() Time {
	return Time{time.Time{}}
}

// TimeFrom creates a new Time wrapping time.Time.
func TimeFrom(t time.Time) Time {
	return Time{t}
}
