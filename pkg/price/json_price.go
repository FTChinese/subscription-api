package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONPrice wrap Price so that it could be saved
// as a JSON column in sql, and marshalled into null for
// empty value when used as introductory.
type JSONPrice struct {
	Price
}

// MarshalJSON encodes an optional price to nullable result.
func (p JSONPrice) MarshalJSON() ([]byte, error) {
	if p.ID == "" {
		return []byte("null"), nil
	}

	return json.Marshal(p.Price)
}

// UnmarshalJSON parses a nullable value to price.
func (p *JSONPrice) UnmarshalJSON(b []byte) error {
	var v Price
	if b == nil {
		*p = JSONPrice{}
		return nil
	}

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	*p = JSONPrice{Price: v}
	return nil
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p JSONPrice) Value() (driver.Value, error) {
	if p.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (p *JSONPrice) Scan(src interface{}) error {
	if src == nil {
		*p = JSONPrice{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp JSONPrice
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to JSONPrice")
	}
}
