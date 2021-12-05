package apple

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"strings"
)

type Environment int

const (
	EnvNull Environment = iota
	EnvProduction
	EnvSandbox
)

var envNames = [...]string{
	"",
	"Production",
	"Sandbox",
}

var envMap = map[Environment]string{
	1: envNames[1],
	2: envNames[2],
}

var envValue = map[string]Environment{
	envNames[1]:  EnvProduction,
	envNames[2]:  EnvSandbox,
	"PROD":       EnvProduction, // Handle Apple's erratic naming convention. It appears in its server-to-server notification.
	"production": EnvProduction,
	"sandbox":    EnvSandbox,
}

// ParseEnvironment turns a string to an Environment instance.
// BUG: In SQL definition we used accidentally ENUM('production', 'sandbox')
// and somehow the SQL's scan ignored case. When parsing, we should
// handle the casing issue.
// To avoid this issue we always turns name to lower case.
func ParseEnvironment(name string) (Environment, error) {
	name = strings.ToLower(name)
	if x, ok := envValue[name]; ok {
		return x, nil
	}

	return EnvNull, fmt.Errorf("%s is not a valid Environment", name)
}

// String always uses capitalized version
func (x Environment) String() string {
	if s, ok := envMap[x]; ok {
		return s
	}

	return ""
}

func (x *Environment) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseEnvironment(s)

	*x = tmp

	return nil
}

func (x Environment) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

func (x *Environment) Scan(src interface{}) error {
	if src == nil {
		*x = EnvNull
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseEnvironment(string(s))
		*x = tmp
		return nil

	default:
		return enum.ErrIncompatible
	}
}

// Value always uses lower case.
func (x Environment) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return strings.ToLower(s), nil
}
