package wechat

import (
	"encoding/json"
	"github.com/objcoding/wxpay"
)

// NativeAppParams is used by native app to call SDK.
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
type NativeAppParams struct {
	AppID     string `json:"appId" map:"appid"`
	Nonce     string `json:"nonce" map:"noncestr"`
	Package   string `json:"pkg" map:"package"`
	PartnerID string `json:"partnerId" map:"partnerid"`
	PrepayID  string `json:"prepayId" map:"prepayid"`
	Signature string `json:"signature" map:"-"`
	Timestamp string `json:"timestamp" map:"timestamp"`
}

func NewNativeAppParams(or OrderResult) NativeAppParams {
	return NativeAppParams{
		AppID:     or.AppID,
		PartnerID: or.MchID,
		PrepayID:  or.PrepayID,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "Sign=WXPay",
	}
}

func (p NativeAppParams) IsZero() bool {
	return p.AppID == ""
}

func (p NativeAppParams) Marshal() wxpay.Params {
	return make(wxpay.Params).
		SetString("appid", p.AppID).
		SetString("partnerid", p.PartnerID).
		SetString("prepayid", p.PrepayID).
		SetString("package", p.Package).
		SetString("noncestr", p.Nonce).
		SetString("timestamp", p.Timestamp)
}

// NativeAppParamsJSON is used to turn an empty NativeAppParams to json null.
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
