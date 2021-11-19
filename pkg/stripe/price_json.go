package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// PriceJSON saves price to or retrieves from db as JSON field.
type PriceJSON struct {
	Price
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (j PriceJSON) Value() (driver.Value, error) {
	if j.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (j *PriceJSON) Scan(src interface{}) error {
	if src == nil {
		*j = PriceJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp PriceJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*j = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to PriceJSON")
	}
}
