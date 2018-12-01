package member

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	standard = "standard"
	premium  = "premium"
)

var tiersRaw = [...]string{
	standard,
	premium,
}

var tiersCN = [...]string{
	"标准会员",
	"高级会员",
}

var tiersEN = [...]string{
	"Standard",
	"Premium",
}

// Tier is an enum.
type Tier int

// UnmarshalJSON implements the Unmarshaler interface.
func (t *Tier) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tier, ok := tierEnum[s]

	if !ok {
		return errors.New("only standard and premium member tier allowed")
	}

	*t = tier

	return nil
}

// MarshalJSON impeoments the Marshaler interface
func (t Tier) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
func (t *Tier) Scan(src interface{}) error {
	var source string
	switch src.(type) {
	case string:
		source = src.(string)

	default:
		return errors.New("incompatible type for member tier")
	}

	tier, err := NewTier(source)
	if err != nil {
		return err
	}

	*t = tier

	return nil
}

// Value implements driver.Valuer interface to save value into SQL.
func (t Tier) Value() (driver.Value, error) {
	s := t.String()
	if s == "" {
		return driver.Value(""), fmt.Errorf("member tier %d is not a valid member type", t)
	}

	return driver.Value(s), nil
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
	TierFree     Tier = -1
	TierStandard Tier = 0
	TierPremium  Tier = 1
)

// Maps raw value to Tier type.
var tierEnum = map[string]Tier{
	standard: TierStandard,
	premium:  TierPremium,
}

// NewTier converts a string into a MemberTier type.
func NewTier(key string) (Tier, error) {
	tier, ok := tierEnum[key]

	if !ok {
		return TierFree, errors.New("Only standard and premium tier allowed")
	}

	return tier, nil
}
