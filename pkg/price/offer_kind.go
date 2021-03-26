package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type OfferKind string

const (
	OfferKindNull      OfferKind = ""
	OfferKindPromotion OfferKind = "promotion"
	OfferKindRetention OfferKind = "retention"
	OfferKindWinBack   OfferKind = "win_back"
)

func (x *OfferKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*x = OfferKind(s)

	return nil
}

func (x OfferKind) MarshalJSON() ([]byte, error) {
	if x == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + x + `"`), nil
}

func (x *OfferKind) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = OfferKind(s)
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x OfferKind) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}

func (x OfferKind) ContainedBy(kinds []OfferKind) bool {
	for _, v := range kinds {
		if v == x {
			return true
		}
	}

	return false
}
