package sq

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type StringList []string

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (l StringList) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (l *StringList) Scan(src interface{}) error {
	if src == nil {
		*l = StringList{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp []string
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}
