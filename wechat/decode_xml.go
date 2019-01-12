package wechat

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/view"
)

// DecodeXML parses wxpay's weird response XML data.
// See https://github.com/objcoding/wxpay/issues/10
func DecodeXML(r io.Reader) wxpay.Params {
	var (
		d      *xml.Decoder
		start  *xml.StartElement
		params wxpay.Params
	)
	d = xml.NewDecoder(r)
	params = make(wxpay.Params)
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			start = &t
		case xml.CharData:
			if t = bytes.TrimSpace(t); len(t) > 0 {
				params.SetString(start.Name.Local, string(t))
			}
		}
	}
	return params
}

// ValidateResponse verifies if wechat response is valid.
//
// Example response:
// return_code:SUCCESS|FAIL
// return_msg:OK
//
// Present only if return_code == SUCCESS
// appid:wx......
// mch_id:........
// nonce_str:8p8ZlUBkLsFPxC6g
// sign:DB68F0D9F193D499DF9A2EDBFFEAF312
// result_code:SUCCESS|FAIL
// err_code
// err_code_des
//
// Present only if returnd_code == SUCCESS and result_code == SUCCCESS
// trade_type:APP
// prepay_id:wx20125006086590be8d9519f40090763508
// NOTE: this sdk treat return_code == FAIL as valid.
// Possible return_msg:
// appid不存在;
// 商户号mch_id与appid不匹配;
// invalid spbill_create_ip;
// spbill_create_ip参数长度有误; (Wx does not accept IPv6 like 9b5b:2ef9:6c9f:cf5:130e:984d:8958:75f9 :-<)
func ValidateResponse(resp wxpay.Params) *view.Reason {
	if resp.GetString("return_code") == wxpay.Fail {
		returnMsg := resp.GetString("return_msg")
		logger.
			WithField("trace", "ValidateResponse").
			Errorf("return_code is FAIL. return_msg: %s", returnMsg)

		reason := &view.Reason{
			Field: "return_code",
			Code:  "fail",
		}
		reason.SetMessage(returnMsg)

		return reason
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("trace", "ValidateResponse").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")

		reason := &view.Reason{
			Field: "result_code",
			Code:  errCode,
		}
		reason.SetMessage(errCodeDes)

		return reason
	}

	return nil
}
