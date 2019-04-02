package wechat

type TradeType int

const (
	TradeTypeDesktop TradeType = iota
	TradeTypeMobile
	TradeTypeJSAPI
	TradeTypeApp
)

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
