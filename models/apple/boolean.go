package apple

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Boolean int

const (
	BooleanNul Boolean = iota
	BooleanFalse
	BooleanTrue
)

var booleanStrings = [...]string{
	"",
	"false",
	"true",
}

func ParseBoolean(name string) (Boolean, error) {
	switch name {
	case booleanStrings[0]:
		return BooleanFalse, nil

	case booleanStrings[1]:
		return BooleanTrue, nil

	default:
		return BooleanNul, fmt.Errorf("%s is not a valid Boolean", name)
	}
}

func (x Boolean) String() string {
	if x < BooleanNul || x > BooleanTrue {
		return ""
	}

	return booleanStrings[x]
}

func (x *Boolean) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseBoolean(s)

	*x = tmp

	return nil
}

func (x Boolean) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}
