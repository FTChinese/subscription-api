package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ColumnMembership struct {
	Membership
}

// Value implements Valuer interface by saving the entire
// type as JSON string, or null if it is a zero value.
func (m ColumnMembership) Value() (driver.Value, error) {
	// For zero value, save as NULL.
	if m.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (m *ColumnMembership) Scan(src interface{}) error {
	// Handle null value.
	if src == nil {
		*m = ColumnMembership{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnMembership
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*m = tmp
		return nil

	default:
		return errors.New("incompatible type to scna to ColumnMembership")
	}
}
