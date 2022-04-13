package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// FtcPriceJSON wrap Price so that it could be saved
// as a JSON column in sql, and marshalled into null for
// empty value when used as introductory.
type FtcPriceJSON struct {
	FtcPrice
}

// MarshalJSON encodes an optional price to nullable result.
func (p FtcPriceJSON) MarshalJSON() ([]byte, error) {
	if p.ID == "" {
		return []byte("null"), nil
	}

	return json.Marshal(p.FtcPrice)
}

// UnmarshalJSON parses a nullable value to price.
func (p *FtcPriceJSON) UnmarshalJSON(b []byte) error {
	var v FtcPrice
	if b == nil {
		*p = FtcPriceJSON{}
		return nil
	}

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	*p = FtcPriceJSON{FtcPrice: v}
	return nil
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p FtcPriceJSON) Value() (driver.Value, error) {
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
func (p *FtcPriceJSON) Scan(src interface{}) error {
	if src == nil {
		*p = FtcPriceJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp FtcPriceJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to FtcPriceJSON")
	}
}
