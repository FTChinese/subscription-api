package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ColumnDiscount is used to save/retrieve Discount in a SQL JSON column.
type ColumnDiscount struct {
	Discount
}

func (d ColumnDiscount) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(d.Discount)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (d *ColumnDiscount) Scan(src interface{}) error {
	if src == nil {
		*d = ColumnDiscount{}
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnDiscount
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*d = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to JSONPrice")
	}
}
