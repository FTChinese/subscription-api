package plan

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

type SubsKind int

const (
	SubsKindNull SubsKind = iota // If user remaining subs period exceed max allowable one, or any other error.
	SubsKindCreate
	SubsKindRenew
	SubsKindUpgrade
)

var subsKindNames = [...]string{
	"",
	"create",
	"renew",
	"upgrade",
}

var subsKindMap = map[SubsKind]string{
	1: subsKindNames[1],
	2: subsKindNames[2],
	3: subsKindNames[3],
}

var subsKindValue = map[string]SubsKind{
	subsKindNames[1]: 1,
	subsKindNames[2]: 2,
	subsKindNames[3]: 3,
}

func ParseSubsKind(name string) (SubsKind, error) {
	if x, ok := subsKindValue[name]; ok {
		return x, nil
	}

	return SubsKindNull, fmt.Errorf("%s is not valid SubsKind", name)
}

func (x SubsKind) String() string {
	if s, ok := subsKindMap[x]; ok {
		return s
	}
	return ""
}

func (x SubsKind) StringCN() string {
	switch x {
	case SubsKindCreate:
		return  "订阅"

	case SubsKindRenew:
		return "续订"

	case SubsKindUpgrade:
		return "升级订阅"

	default:
		return ""
	}
}

// SnapshotReason converts to SnapshotReason for renew or upgrade.
func (x SubsKind) SnapshotReason() enum.SnapshotReason {
	switch x {
	case SubsKindRenew:
		return enum.SnapshotReasonRenew

	case SubsKindUpgrade:
		return enum.SnapshotReasonUpgrade

	default:
		return enum.SnapshotReasonNull
	}
}
func (x *SubsKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseSubsKind(s)

	*x = tmp

	return nil
}

func (x SubsKind) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

func (x *SubsKind) Scan(src interface{}) error {
	if src == nil {
		*x = SubsKindNull
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseSubsKind(string(s))
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x SubsKind) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}
