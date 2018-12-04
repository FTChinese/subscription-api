package member

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gitlab.com/ftchinese/subscription-api/util"
)

const (
	alipay = "alipay"
	wxpay  = "tenpay"
	stripe = "stripe"
)

var payRaw = [...]string{
	alipay,
	wxpay,
	stripe,
}

var payCN = [...]string{
	"支付宝",
	"微信支付",
	"Stripe",
}

var payEN = [...]string{
	"Zhifubao",
	"Wechat Pay",
	"Stripe",
}

// PayMethod is an enum for payment methods
type PayMethod int

// UnmarshalJSON implements the Unmarshaler interface.
func (m *PayMethod) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	method, err := NewPayMethod(s)

	if err != nil {
		return err
	}

	*m = method

	return nil
}

// MarshalJSON impeoments the Marshaler interface
func (m PayMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// Scan implements sql.Scanner interface to retrieve value from SQL.
// SQL null will be turned into zero value InvalidPay.
func (m *PayMethod) Scan(src interface{}) error {
	if src == nil {
		*m = InvalidPay
		return nil
	}

	switch s := src.(type) {
	case []byte:
		method, err := NewPayMethod(string(s))
		if err != nil {
			return err
		}
		*m = method
		return nil

	default:
		return util.ErrIncompatible
	}
}

// Value implements driver.Valuer interface to save value into SQL.
func (m PayMethod) Value() (driver.Value, error) {
	s := m.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}

func (m PayMethod) String() string {
	if m < Alipay || m > Stripe {
		return ""
	}

	return payRaw[m]
}

// ToCN output cycle as Chinese text
func (m PayMethod) ToCN() string {
	if m < Alipay || m > Stripe {
		return ""
	}

	return payCN[m]
}

// ToEN output cycle as English text
func (m PayMethod) ToEN() string {
	if m < Alipay || m > Stripe {
		return ""
	}

	return payEN[m]
}

// Supported payment methods
const (
	InvalidPay PayMethod = -1
	Alipay     PayMethod = 0
	Wxpay      PayMethod = 1
	Stripe     PayMethod = 2
)

// NewPayMethod creates a new instance of PayMethod
func NewPayMethod(method string) (PayMethod, error) {
	switch method {
	case alipay:
		return Alipay, nil
	case wxpay:
		return Wxpay, nil
	case stripe:
		return Stripe, nil
	default:
		return InvalidPay, errors.New("Raw value for payment method could only be alipay, tenpay, or stripe")
	}
}
