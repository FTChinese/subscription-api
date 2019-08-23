package util

import (
	"database/sql/driver"
	"errors"
	"strings"
)

type StringSlice []string

func (x *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*x = []string{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp := strings.Split(string(s), ",")
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x StringSlice) Value() (driver.Value, error) {
	s := strings.Join(x, ",")
	if s == "" {
		return nil, nil
	}

	return s, nil
}
