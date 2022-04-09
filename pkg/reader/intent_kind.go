package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

type SubsIntentKind int

const (
	IntentNull SubsIntentKind = iota
	IntentCreate
	IntentRenew
	IntentUpgrade // Switching subscription tier, e.g., from standard to premium.
	IntentDowngrade
	IntentAddOn
	IntentOneTimeToAutoRenew // Same as new, with valid remaining membership period.
	IntentSwitchInterval     // Switching subscription billing cycle, e.g., from month to year.
	IntentForbidden
)

var intentKindNames = []string{
	"",
	"create",
	"renew",
	"upgrade",
	"downgrade",
	"addon",
	"one_time_to_auto_renew",
	"switch_interval",
	"forbidden",
}

var subsKindMap = map[SubsIntentKind]string{
	1: intentKindNames[1],
	2: intentKindNames[2],
	3: intentKindNames[3],
	4: intentKindNames[4],
	5: intentKindNames[5],
	6: intentKindNames[6],
	7: intentKindNames[7],
	8: intentKindNames[8],
}

var subsKindValue = map[string]SubsIntentKind{
	intentKindNames[1]: 1,
	intentKindNames[2]: 2,
	intentKindNames[3]: 3,
	intentKindNames[4]: 4,
	intentKindNames[5]: 5,
	intentKindNames[6]: 6,
	intentKindNames[7]: 7,
	intentKindNames[8]: 8,
}

func ParseSubsKind(name string) (SubsIntentKind, error) {
	if x, ok := subsKindValue[name]; ok {
		return x, nil
	}

	return IntentForbidden, fmt.Errorf("%s is not valid SubsIntentKind", name)
}

func (x SubsIntentKind) String() string {
	if s, ok := subsKindMap[x]; ok {
		return s
	}

	return ""
}

// UnmarshalJSON implements the Unmarshaler interface.
func (x *SubsIntentKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseSubsKind(s)

	*x = tmp

	return nil
}

// MarshalJSON implements the Marshaler interface
func (x SubsIntentKind) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
// SQL null will be turned into zero value TierFree.
func (x *SubsIntentKind) Scan(src interface{}) error {
	if src == nil {
		*x = IntentForbidden
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseSubsKind(string(s))
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to SubsIntentKind")
	}
}

// Value implements driver.Valuer interface to save value into SQL.
func (x SubsIntentKind) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

func (x SubsIntentKind) IsNewSubs() bool {
	return x == IntentCreate || x == IntentOneTimeToAutoRenew
}

func (x SubsIntentKind) IsUpdating() bool {
	return x == IntentUpgrade || x == IntentSwitchInterval
}

func (x SubsIntentKind) IsSwitchToAutoRenew() bool {
	return x == IntentOneTimeToAutoRenew
}

func (x SubsIntentKind) IsForbidden() bool {
	return x == IntentForbidden
}

var intentToOrderKind = map[SubsIntentKind]enum.OrderKind{
	IntentCreate:    enum.OrderKindCreate,
	IntentRenew:     enum.OrderKindRenew,
	IntentUpgrade:   enum.OrderKindUpgrade,
	IntentDowngrade: enum.OrderKindDowngrade,
	IntentAddOn:     enum.OrderKindAddOn,
}

func (x SubsIntentKind) ToOrderKind() enum.OrderKind {
	return intentToOrderKind[x]
}
