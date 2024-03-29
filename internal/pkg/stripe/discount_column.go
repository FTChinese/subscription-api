package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// DiscountColumn is used to save discount column under subscription.
type DiscountColumn struct {
	Discount
}

func (x DiscountColumn) MarshalJSON() ([]byte, error) {
	if x.ID == "" {
		return []byte("null"), nil
	}

	return json.Marshal(x.Discount)
}

func (x *DiscountColumn) UnmarshalJSON(b []byte) error {
	var d Discount
	if b == nil {
		*x = DiscountColumn{}
		return nil
	}

	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}

	*x = DiscountColumn{d}

	return nil
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (x DiscountColumn) Value() (driver.Value, error) {
	if x.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(x.Discount)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (x *DiscountColumn) Scan(src interface{}) error {
	if src == nil {
		*x = DiscountColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp Discount
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*x = DiscountColumn{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to DiscountColumn")
	}
}
