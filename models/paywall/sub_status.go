package paywall

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type SubStatus int

const (
	SubStatusNull SubStatus = iota
	SubStatusActive
	SubStatusCanceled
	SubStatusIncomplete
	SubStatusIncompleteExpired
	SubStatusPastDue
	SubStatusTrialing
	SubStatusUnpaid
)

var subStatusNames = [...]string{
	"",
	"active",
	"canceled",
	"incomplete",
	"incomplete_expired",
	"past_due",
	"trialing",
	"unpaid",
}

// Map SubStatus to string value to be persisted.
var subStatusMap = map[SubStatus]string{
	1: subStatusNames[1],
	2: subStatusNames[2],
	3: subStatusNames[3],
	4: subStatusNames[4],
	5: subStatusNames[5],
	6: subStatusNames[6],
	7: subStatusNames[7],
}

// Parse a string to SubStatus
var subStatusValue = map[string]SubStatus{
	subStatusNames[1]: 1,
	subStatusNames[2]: 2,
	subStatusNames[3]: 3,
	subStatusNames[4]: 4,
	subStatusNames[5]: 5,
	subStatusNames[6]: 6,
	subStatusNames[7]: 7,
}

// ParseSubStatus turns a string to SubStatus.
func ParseSubStatus(name string) (SubStatus, error) {
	if x, ok := subStatusValue[name]; ok {
		return x, nil
	}

	return SubStatusNull, fmt.Errorf("%s is not valid SubStatus", name)
}

// ShouldCreate checks whether membership's current status
// should allow creation of a new membership.
func (x SubStatus) ShouldCreate() bool {
	return x == SubStatusNull ||
		x == SubStatusIncompleteExpired ||
		x == SubStatusPastDue ||
		x == SubStatusCanceled ||
		x == SubStatusUnpaid
}

func (x SubStatus) String() string {
	if s, ok := subStatusMap[x]; ok {
		return s
	}

	return ""
}

func (x *SubStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseSubStatus(s)

	*x = tmp

	return nil
}

func (x SubStatus) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`""` + s + `""`), nil
}

func (x *SubStatus) Scan(src interface{}) error {
	if src == nil {
		*x = SubStatusNull
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseSubStatus(string(s))
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x SubStatus) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}
