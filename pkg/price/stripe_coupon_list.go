package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type StripeCouponList []StripeCoupon

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (c StripeCouponList) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (c *StripeCouponList) Scan(src interface{}) error {
	if src == nil {
		*c = StripeCouponList{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp []StripeCoupon
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*c = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to StripeCouponList")
	}
}
