// +build !production

package wechat

import "github.com/objcoding/wxpay"

func (r BaseResp) MockToMap() wxpay.Params {
	p := make(wxpay.Params)

	p.SetString("return_code", r.ReturnCode)
	p.SetString("return_msg", r.ReturnMessage.String)
	p.SetString("appid", r.AppID.String)
	p.SetString("mch_id", r.MID.String)
	p.SetString("nonce_str", r.Nonce.String)
	p.SetString("result_code", r.ResultCode.String)

	return p
}

func (or OrderResp) MockToMap() wxpay.Params {
	p := or.BaseResp.MockToMap()
	p.SetString("prepay_id", or.PrepayID.String)

	return p
}

func (n Notification) MockToMap() wxpay.Params {
	p := n.BaseResp.MockToMap()

	var subscribed string
	if n.IsSubscribed {
		subscribed = "Y"
	} else {
		subscribed = "N"
	}

	if n.OpenID.Valid {
		p.SetString("openid", n.OpenID.String)
	}

	p.SetString("is_subscribe", subscribed)
	p.SetString("bank_type", n.BankType.String)
	p.SetInt64("total_fee", n.TotalFee.Int64)
	p.SetString("transaction_id", n.TransactionID.String)
	p.SetString("out_trade_no", n.FTCOrderID.String)
	p.SetString("time_end", n.TimeEnd.String)

	return p
}
