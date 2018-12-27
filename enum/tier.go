package enum

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	tierFree = ""
	standard = "standard"
	premium  = "premium"
)

var tiersRaw = [...]string{
	tierFree,
	standard,
	premium,
}

var tiersCN = [...]string{
	"",
	"标准会员",
	"高级会员",
}

var tiersEN = [...]string{
	"",
	"Standard",
	"Premium",
}

// Tier is an enum.
type Tier int

// IsValid tests if t is one of TierStandard or TierPremium
func (t Tier) IsValid() bool {
	return t != TierFree
}

// UnmarshalJSON implements the Unmarshaler interface.
func (t *Tier) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*t = NewTier(s)

	return nil
}

// MarshalJSON impeoments the Marshaler interface
func (t Tier) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
// SQL null will be turned into zero value TierFree.
func (t *Tier) Scan(src interface{}) error {
	if src == nil {
		*t = TierFree
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*t = NewTier(string(s))
		return nil

	default:
		return ErrIncompatible
	}
}

// Value implements driver.Valuer interface to save value into SQL.
func (t Tier) Value() (driver.Value, error) {
	s := t.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

func (t Tier) String() string {
	if t < TierStandard || t > TierPremium {
		return ""
	}

	return tiersRaw[t]
}

// ToCN output tier as Chinese text
func (t Tier) ToCN() string {
	if t < TierStandard || t > TierPremium {
		return ""
	}

	return tiersCN[t]
}

// ToEN output tier as English text
func (t Tier) ToEN() string {
	if t < TierStandard || t > TierPremium {
		return ""
	}

	return tiersEN[t]
}

// Values of MemberTier
const (
	TierFree     Tier = 0
	TierStandard Tier = 1
	TierPremium  Tier = 2
)

// NewTier converts a string into a MemberTier type.
func NewTier(tier string) Tier {
	switch tier {
	case standard:
		return TierStandard
	case premium:
		return TierPremium
	default:
		return TierFree
	}
}
