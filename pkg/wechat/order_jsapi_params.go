package wechat

import (
	"encoding/json"
	"github.com/objcoding/wxpay"
)

// JSApiParams creates an order when user is trying to pay
// inside wechat's embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type JSApiParams struct {
	AppID     string `json:"appId"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	SignType  string `json:"signType"`
	Signature string `json:"signature"` // This could only be generated after ToMap is called.
	Timestamp string `json:"timestamp"`
}

// NewJSApiParams creates a new JSApiParams to be signed.
func NewJSApiParams(or OrderResult) JSApiParams {
	return JSApiParams{
		AppID:     or.AppID,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "prepay_id=" + or.PrepayID,
		SignType:  "MD5",
	}
}

func (p JSApiParams) IsZero() bool {
	return p.Signature == ""
}

// ToMap turns struct to a map so that we could generate signature from sdk.
func (p JSApiParams) ToMap() wxpay.Params {
	return wxpay.Params{
		"appId":     p.AppID,
		"timeStamp": p.Timestamp,
		"nonceStr":  p.Nonce,
		"package":   p.Package,
		"signType":  p.SignType,
	}
}

// JSApiParamsJSON is used to implement Marshaller interface
// to avoid cyclic call of MarshalJSON.
type JSApiParamsJSON struct {
	JSApiParams
}

func (p JSApiParamsJSON) MarshalJSON() ([]byte, error) {
	if p.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(p.JSApiParams)
}

// UnmarshalJSON parses a nullable value to price.
func (p *JSApiParamsJSON) UnmarshalJSON(b []byte) error {

	if b == nil {
		*p = JSApiParamsJSON{}
		return nil
	}

	var v JSApiParams

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	*p = JSApiParamsJSON{JSApiParams: v}
	return nil
}
