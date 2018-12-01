package member

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	year  = "year"
	month = "month"
)

var cyclesRaw = [...]string{
	year,
	month,
}

// Chinese translation
var cyclesCN = [...]string{
	"年",
	"月",
}

// English translation
var cyclesEN = [...]string{
	"Year",
	"Month",
}

// Cycle is an enum for billing cycles.
type Cycle int

// TimeAfterACycle adds one cycle to a time instance and returns the new time.
func (c Cycle) TimeAfterACycle(t time.Time) (time.Time, error) {
	switch c {
	case CycleYear:
		return t.AddDate(1, 0, 1), nil
	case CycleMonth:
		return t.AddDate(0, 1, 1), nil
	default:
		return t, errors.New("not a valid cycle type")
	}
}

// UnmarshalJSON implements the Unmarshaler interface.
func (c *Cycle) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	cycle, err := NewCycle(s)

	if err != nil {
		return err
	}

	*c = cycle

	return nil
}

// MarshalJSON impeoments the Marshaler interface
func (c Cycle) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
func (c *Cycle) Scan(src interface{}) error {
	if src == nil {
		*c = CycleInvalid
		return nil
	}

	switch s := src.(type) {
	case []byte:
		cycle, err := NewCycle(string(s))
		if err != nil {
			return err
		}
		*c = cycle
		return nil

	default:
		return ErrIncompatible
	}
}

// Value implements driver.Valuer interface to save value into SQL.
func (c Cycle) Value() (driver.Value, error) {
	s := c.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

func (c Cycle) String() string {
	if c < CycleYear || c > CycleMonth {
		return ""
	}

	return cyclesRaw[c]
}

// ToCN output cycle as Chinese text
func (c Cycle) ToCN() string {
	if c < CycleYear || c > CycleMonth {
		return ""
	}

	return cyclesCN[c]
}

// ToEN output cycle as English text
func (c Cycle) ToEN() string {
	if c < CycleYear || c > CycleMonth {
		return ""
	}

	return cyclesEN[c]
}

// Supported billing cycles
const (
	CycleInvalid Cycle = -1
	CycleYear    Cycle = 0
	CycleMonth   Cycle = 1
)

// NewCycle creates a new instance of Cycle.
func NewCycle(key string) (Cycle, error) {
	switch key {
	case year:
		return CycleYear, nil
	case month:
		return CycleMonth, nil
	default:
		return CycleInvalid, errors.New("Only year and month billing cycle allowed")
	}
}
