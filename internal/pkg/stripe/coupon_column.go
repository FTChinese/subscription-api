package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type CouponColumn struct {
	price.StripeCoupon
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (x CouponColumn) Value() (driver.Value, error) {
	if x.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(x.StripeCoupon)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (x *CouponColumn) Scan(src interface{}) error {
	if src == nil {
		*x = CouponColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp price.StripeCoupon
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*x = CouponColumn{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to CouponColumn")
	}
}
