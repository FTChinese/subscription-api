package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type DiscountStatus string

const (
	DiscountStatusNull      DiscountStatus = ""
	DiscountStatusActive    DiscountStatus = "active"
	DiscountStatusPaused    DiscountStatus = "paused"
	DiscountStatusCancelled DiscountStatus = "cancelled"
)

func (x *DiscountStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*x = DiscountStatus(s)

	return nil
}

func (x DiscountStatus) MarshalJSON() ([]byte, error) {
	if x == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + x + `"`), nil
}

func (x *DiscountStatus) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = DiscountStatus(s)
		return nil

	default:
		return errors.New("incompatible type to scan to DiscountStatus")
	}
}

func (x DiscountStatus) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
