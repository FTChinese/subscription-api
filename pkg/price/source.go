package price

import (
	"database/sql/driver"
	"errors"
)

type Source string

const (
	SourceUnknown Source = ""
	SourceFTC     Source = "ftc"
	SourceStripe  Source = "stripe"
)

func (x *Source) Scan(src interface{}) error {
	if src == nil {
		*x = SourceUnknown
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = Source(s)
		return nil

	default:
		return errors.New("incompatible type to scan to Source")
	}
}

func (x Source) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
