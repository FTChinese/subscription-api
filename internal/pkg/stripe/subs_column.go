package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type SubsColumn struct {
	Subs
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (x SubsColumn) Value() (driver.Value, error) {
	if x.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(x.Subs)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (x *SubsColumn) Scan(src interface{}) error {
	if src == nil {
		*x = SubsColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp Subs
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*x = SubsColumn{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to SubsColumn")
	}
}
