package addon

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Source string

const (
	SourceUpgradeCarryOver      Source = "upgrade_carry_over"
	SourceOneTimeToSubCarryOver Source = "one_time_to_sub_carry_over"
	SourceCompensation          Source = "compensation"
	SourceUserPurchase          Source = "user_purchase"
)

func (x *Source) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*x = Source(s)

	return nil
}

func (x Source) MarshalJSON() ([]byte, error) {
	if x == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + x + `"`), nil
}

func (x *Source) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = Source(s)
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x Source) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
