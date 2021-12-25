package wechat

// TradeType is an enum for wechat's various nuance payment
// channel.
type TradeType string

const (
	TradeTypeDesktop TradeType = "NATIVE"
	TradeTypeMobile  TradeType = "MWEB"
	TradeTypeJSAPI   TradeType = "JSAPI"
	TradeTypeApp     TradeType = "APP"
)
