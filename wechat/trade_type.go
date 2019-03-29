package wechat

type TradeType int

const (
	TradeTypeWeb TradeType = iota
	TradeTypeApp
	TradeTypeJSAPI
)

func (x TradeType) String() string {
	names := [...]string{
		"NATIVE",
		"APP",
		"JSAPI",
	}

	if x < TradeTypeWeb || x > TradeTypeJSAPI {
		return ""
	}

	return names[x]
}
