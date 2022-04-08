package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type SubsKind int

// TODO: move to pw package.
const (
	SubsKindForbidden SubsKind = iota
	SubsKindCreate
	SubsKindRenew
	SubsKindUpgrade // Switching subscription tier, e.g., from standard to premium.
	SubsKindDowngrade
	SubsKindAddOn
	SubsKindOneTimeToAutoRenew // Same as new, with valid remaining membership period.
	SubsKindSwitchInterval     // Switching subscription billing cycle, e.g., from month to year.
	SubsKindRefreshAutoRenew
)

var subsKindNames = []string{
	"",
	"create",
	"renew",
	"upgrade",
	"downgrade",
	"addon",
	"one_time_to_auto_renew",
	"switch_interval",
	"refresh_auto_renew",
}

var subsKindMap = map[SubsKind]string{
	1: subsKindNames[1],
	2: subsKindNames[2],
	3: subsKindNames[3],
	4: subsKindNames[4],
	5: subsKindNames[5],
	6: subsKindNames[6],
	7: subsKindNames[7],
	8: subsKindNames[8],
}

var subsKindValue = map[string]SubsKind{
	subsKindNames[1]: 1,
	subsKindNames[2]: 2,
	subsKindNames[3]: 3,
	subsKindNames[4]: 4,
	subsKindNames[5]: 5,
	subsKindNames[6]: 6,
	subsKindNames[7]: 7,
	subsKindNames[8]: 8,
}

func ParseSubsKind(name string) (SubsKind, error) {
	if x, ok := subsKindValue[name]; ok {
		return x, nil
	}

	return SubsKindForbidden, fmt.Errorf("%s is not valid SubsKind", name)
}

func (x SubsKind) String() string {
	if s, ok := subsKindMap[x]; ok {
		return s
	}

	return ""
}

// UnmarshalJSON implements the Unmarshaler interface.
func (x *SubsKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseSubsKind(s)

	*x = tmp

	return nil
}

// MarshalJSON implements the Marshaler interface
func (x SubsKind) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
// SQL null will be turned into zero value TierFree.
func (x *SubsKind) Scan(src interface{}) error {
	if src == nil {
		*x = SubsKindForbidden
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseSubsKind(string(s))
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to SubsKind")
	}
}

// Value implements driver.Valuer interface to save value into SQL.
func (x SubsKind) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

func (x SubsKind) IsNewSubs() bool {
	return x == SubsKindCreate || x == SubsKindOneTimeToAutoRenew
}

func (x SubsKind) IsUpdating() bool {
	return x == SubsKindUpgrade || x == SubsKindSwitchInterval
}
