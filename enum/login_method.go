package enum

import (
	"database/sql/driver"
	"fmt"
)

// LoginMethod is an enumeration of login method.
type LoginMethod int

// InvalidLogin represents an unknown logoin method
const InvalidLogin LoginMethod = -1

// Allowed values for LoginMethod
const (
	EmailLogin LoginMethod = iota
	WechatLogin
)

var loginMethodNames = [...]string{
	"email",
	"wechat",
}

var loginMethodMap = map[LoginMethod]string{
	0: loginMethodNames[0],
	1: loginMethodNames[1],
}

var loginMethodValue = map[string]LoginMethod{
	loginMethodNames[0]: 0,
	loginMethodNames[1]: 1,
}

// ParseLoginMethod creates a new LoginMethod from a string: email or wechat.
func ParseLoginMethod(name string) (LoginMethod, error) {
	if x, ok := loginMethodValue[name]; ok {
		return x, nil
	}

	return InvalidLogin, fmt.Errorf("%s is not a valid LoginMethod", name)
}

func (x LoginMethod) String() string {
	if str, ok := loginMethodMap[x]; ok {
		return str
	}

	return ""
}

// Scan implements the Scanner interface
func (x *LoginMethod) Scan(value interface{}) error {
	var name string
	switch v := value.(type) {
	case string:
		name = v
	case []byte:
		name = string(v)
	case nil:
		*x = InvalidLogin
		return nil
	}

	tmp, err := ParseLoginMethod(name)

	if err != nil {
		return err
	}

	*x = tmp
	return nil
}

// Value implements the Valuer interface.
func (x LoginMethod) Value() (driver.Value, error) {
	if x == InvalidLogin {
		return nil, nil
	}

	return x.String(), nil
}
