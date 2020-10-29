package wechat

// UnifiedOrder contains the data sent to wechat to obtain a payment intent.
type UnifiedOrderConfig struct {
	IP        string
	TradeType TradeType
	OpenID    string // Required only for JSAPI.
}
