package wechat

import (
	"github.com/go-pay/gopay"
	"github.com/objcoding/wxpay"
)

const (
	SignTypeMD5    = "MD5"
	SignTypeSha256 = "HMAC-SHA256"
	Fail           = "FAIL"
	Success        = "SUCCESS"
)

const (
	keyReturnCode   = "return_code"
	keyReturnMsg    = "return_msg"
	keyAppID        = "appid"
	keyMchID        = "mch_id"
	keyBody         = "body"
	keyNonceStr     = "nonce_str"
	keyPrepayID     = "prepay_id"
	keyCodeURL      = "code_url"
	keyMobileWebURL = "mweb_url"
	keyOrderID      = "out_trade_no"
	keyTotalAmount  = "total_fee"
	keyIP           = "spbill_create_ip"
	keyTxnID        = "transaction_id"
	keyEndTime      = "time_end"
	keySign         = "sign"
	keySignType     = "sign_type"
	keyWebhookURL   = "notify_url"
	keyOpenID       = "openid"
	keyTradeType    = "trade_type"
	keyResultCode   = "result_code"
	keyErrCode      = "err_code"
	keyErrCodeDes   = "err_code_des"
)

func GetAppID(bm wxpay.Params) string {
	return bm.GetString(keyAppID)
}

func getMchID(p wxpay.Params) string {
	return p.GetString(keyMchID)
}

func getNonce(p wxpay.Params) string {
	return p.GetString(keyNonceStr)
}

func getSign(bm gopay.BodyMap) string {
	return bm.GetString(keySign)
}

func GetOrderID(p wxpay.Params) string {
	return p.GetString(keyOrderID)
}
