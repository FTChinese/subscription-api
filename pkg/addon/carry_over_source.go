package addon

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type CarryOverSource string

const (
	CarryOverFromUpgrade         CarryOverSource = "one_time_upgrade"
	CarryOverFromSwitchingStripe CarryOverSource = "one_time_to_stripe"
)

func (x *CarryOverSource) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*x = CarryOverSource(s)

	return nil
}

func (x CarryOverSource) MarshalJSON() ([]byte, error) {
	if x == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + x + `"`), nil
}

func (x *CarryOverSource) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = CarryOverSource(s)
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x CarryOverSource) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
