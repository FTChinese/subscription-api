package wxpay

import (
	"encoding/json"
	"github.com/go-pay/gopay/wechat"
)

type NativeAppParams struct {
	AppID     string `json:"appId"`
	PartnerID string `json:"partnerId"`
	PrepayID  string `json:"prepayId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
}

func NewNativeAppParams(resp *wechat.UnifiedOrderResponse) NativeAppParams {
	return NativeAppParams{
		AppID:     resp.Appid,
		PartnerID: resp.MchId,
		PrepayID:  resp.PrepayId,
		Timestamp: GenerateTimestamp(),
		Nonce:     NonceStr(),
		Package:   "Sign=WXPay",
		Signature: "",
	}
}

func (p NativeAppParams) IsZero() bool {
	return p.PrepayID == "" || p.Signature == ""
}

type NativeAppParamsJSON struct {
	NativeAppParams
}

func (p NativeAppParamsJSON) MarshalJSON() ([]byte, error) {
	if p.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(p.NativeAppParams)
}

// UnmarshalJSON parses a nullable value to price.
func (p *NativeAppParamsJSON) UnmarshalJSON(b []byte) error {

	if b == nil {
		*p = NativeAppParamsJSON{}
		return nil
	}

	var v NativeAppParams

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	*p = NativeAppParamsJSON{NativeAppParams: v}
	return nil
}
