package ali

type ProductCode int

const (
	ProductCodeWeb ProductCode = iota
	ProductCodeApp
)

func (x ProductCode) String() string {
	names := [...]string{
		"FAST_INSTANT_TRADE_PAY",
		"QUICK_MSECURITY_PAY",
	}

	if x < ProductCodeWeb || x > ProductCodeApp {
		return ""
	}

	return names[x]
}
