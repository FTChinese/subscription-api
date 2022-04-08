package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// MembershipColumn is used to save Membership as a JSON column
// in MySQL.
type MembershipColumn struct {
	Membership
}

// Value implements Valuer interface by saving the entire
// type as JSON string, or null if it is a zero value.
func (m MembershipColumn) Value() (driver.Value, error) {
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

func (m *MembershipColumn) Scan(src interface{}) error {
	// Handle null value.
	if src == nil {
		*m = MembershipColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp MembershipColumn
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*m = tmp
		return nil

	default:
		return errors.New("incompatible type to scna to MembershipColumn")
	}
}
