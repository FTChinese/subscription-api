package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type OfferKind string

const (
	OfferKindNull         OfferKind = ""
	OfferKindPromotion    OfferKind = "promotion"    // Apply to all uses
	OfferKindRetention    OfferKind = "retention"    // Apply only to valid user
	OfferKindWinBack      OfferKind = "win_back"     // Apply only to expired user
	OfferKindIntroductory OfferKind = "introductory" // Apply only to a new user who has not enjoyed an introductory offer
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
		return errors.New("incompatible type to scan to OfferKind")
	}
}

func (x OfferKind) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}

// ContainedBy checks if this OfferKind is contained by an
// array of OfferKind.
func (x OfferKind) ContainedBy(kinds []OfferKind) bool {
	for _, v := range kinds {
		if v == x {
			return true
		}
	}

	return false
}
