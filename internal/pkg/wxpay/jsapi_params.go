package wxpay

import (
	"encoding/json"
	"github.com/go-pay/gopay/wechat"
)

// JSApiParams creates an order when user is trying to pay
// inside wechat's embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type JSApiParams struct {
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"` // This could only be generated after Marshal is called.
	SignType  string `json:"signType"`
}

func NewJSApiParams(resp *wechat.UnifiedOrderResponse, signType string) JSApiParams {
	return JSApiParams{
		AppID:     resp.Appid,
		Timestamp: GenerateTimestamp(),
		Nonce:     NonceStr(),
		Package:   "prepay_id=" + resp.PrepayId,
		Signature: "",
		SignType:  signType,
	}
}

func (p JSApiParams) IsZero() bool {
	return p.Signature == ""
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
