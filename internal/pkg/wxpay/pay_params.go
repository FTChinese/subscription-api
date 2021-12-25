package wxpay

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/guregu/null"
)

// PayParams contains all parameters required to cal wechat pay
// on every platform.
// Only one of the fields actually have value each time.
// DesktopQr handles payment in desktop browsers.
// wechat send back a custom url for the client to generate a QR image.
// MobileRedirect handles payment in mobile device browser, where you get a canonical url that can be redirected to.
// JsApi fields contains data to perform purchase
// in wechat's embedded browser.
// AppSDK is used by native apps where you embed wechat SDK
// into your app.
type PayParams struct {
	DesktopQr      null.String         `json:"desktopQr"`
	MobileRedirect null.String         `json:"mobileRedirect"`
	JsApi          JSApiParamsJSON     `json:"jsApi"`  // Marshalled to null for empty value.
	AppSDK         NativeAppParamsJSON `json:"appSDK"` // Marshalled to null for empty value.
}

func (p PayParams) IsEmpty() bool {
	return p.DesktopQr.IsZero() && p.MobileRedirect.IsZero() && p.JsApi.IsZero() && p.AppSDK.IsZero()
}

// ColumnPayParams saves SDKParams to a SQL JSON column.
// Empty value is saved as NULL.
type ColumnPayParams struct {
	PayParams
}

func (p ColumnPayParams) Value() (driver.Value, error) {
	if p.IsEmpty() {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (p *ColumnPayParams) Scan(src interface{}) error {
	if src == nil {
		*p = ColumnPayParams{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnPayParams
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to ColumnPayParams")
	}
}
