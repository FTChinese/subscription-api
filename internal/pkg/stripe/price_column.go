package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type PriceColumn struct {
	price.StripePrice
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p PriceColumn) Value() (driver.Value, error) {
	if p.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(p.StripePrice)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (p *PriceColumn) Scan(src interface{}) error {
	if src == nil {
		*p = PriceColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp price.StripePrice
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = PriceColumn{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PriceColumn")
	}
}
