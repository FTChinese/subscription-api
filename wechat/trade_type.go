package wechat

// TradeType is an enum for wechat's various nuance payment
// channel.
type TradeType int

const (
	TradeTypeDesktop TradeType = iota
	TradeTypeMobile
	TradeTypeJSAPI
	TradeTypeApp
)

// String produces values that can be used in request to wechat.
func (x TradeType) String() string {
	names := [...]string{
		"NATIVE",
		"MWEB",
		"JSAPI",
		"APP",
	}

	if x < TradeTypeDesktop || x > TradeTypeApp {
		return ""
	}

	return names[x]
}
