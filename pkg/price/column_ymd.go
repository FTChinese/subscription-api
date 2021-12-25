package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/lib/dt"
)

// ColumnYearMonthDay saves years, months, days in
// a single column as JSON.
type ColumnYearMonthDay struct {
	dt.YearMonthDay
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (j ColumnYearMonthDay) Value() (driver.Value, error) {
	if j.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (j *ColumnYearMonthDay) Scan(src interface{}) error {
	if src == nil {
		*j = ColumnYearMonthDay{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnYearMonthDay
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*j = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to ColumnYearMonthDay")
	}
}
