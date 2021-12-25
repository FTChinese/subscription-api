package wxpay

const (
	SignTypeMD5    = "MD5"
	SignTypeSha256 = "HMAC-SHA256"
	Fail           = "FAIL"
	Success        = "SUCCESS"
)

const (
	keyReturnCode = "return_code"
	keyReturnMsg  = "return_msg"
	keyAppID      = "appid"
	keyMchID      = "mch_id"
	keyNonceStr   = "nonce_str"
	keySign       = "sign"
	keySignTyp    = "sign_type"
	keyResultCode = "result_code"
	keyErrCode    = "err_code"
	keyErrCodeDes = "err_code_des"
)
