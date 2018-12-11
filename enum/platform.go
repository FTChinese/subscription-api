package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// Allowed values for ClientPlatforms
const (
	PlatformNull    ClientPlatform = -1
	PlatformWeb     ClientPlatform = 0
	PlatformIOS     ClientPlatform = 1
	PlatformAndroid ClientPlatform = 2
	pWeb                           = "web"
	piOS                           = "ios"
	pAndroid                       = "android"
)

var platformsRaw = [...]string{
	pWeb,
	piOS,
	pAndroid,
}

// ClientPlatform is used to record on which platoform user is visiting the API.
type ClientPlatform int

func (p ClientPlatform) String() string {
	if !p.IsValid() {
		return ""
	}

	return platformsRaw[p]
}

// IsValid tests if p is one of the valid enums.
func (p ClientPlatform) IsValid() bool {
	if p < PlatformWeb || p > PlatformAndroid {
		return false
	}

	return true
}

// UnmarshalJSON parses a JSON field
func (p *ClientPlatform) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case pWeb:
		*p = PlatformWeb
	case piOS:
		*p = PlatformIOS
	case pAndroid:
		*p = PlatformAndroid
	default:
		*p = PlatformNull
	}

	return nil
}

// MarshalJSON stringifies a ClientPlatform
func (p ClientPlatform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// Scan put a value from SQL.
func (p *ClientPlatform) Scan(src interface{}) error {
	if src == nil {
		*p = PlatformNull
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*p = NewPlatform(string(s))
		return nil

	default:
		*p = PlatformNull
		return nil
	}
}

// Value saves ClientPlatform to SQL ENUM.
func (p ClientPlatform) Value() (driver.Value, error) {
	if !p.IsValid() {
		return nil, nil
	}

	s := p.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

// NewPlatform creates a new instance of ClientPlatform from a string.
func NewPlatform(value string) ClientPlatform {
	switch value {
	case pWeb:
		return PlatformWeb
	case piOS:
		return PlatformIOS
	case pAndroid:
		return PlatformAndroid
	default:
		return PlatformNull
	}
}
