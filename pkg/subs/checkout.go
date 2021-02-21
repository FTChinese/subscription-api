package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func WebhookURL(sandbox bool, method enum.PayMethod) string {
	var baseURL string
	if sandbox {
		baseURL = "http://www.ftacademy.cn/api/sandbox"
	} else {
		baseURL = "http://www.ftacademy.cn/api/v1"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}

// PaymentTitle is used as the value of `subject` for alipay,
// and `body` for wechat pay.
// * 订阅FT中文网标准会员/年
// * 订阅FT中文网高端会员/年
// * 购买FT中文网标准会员/年
// * 购买FT中文网高端会员/年
func PaymentTitle(k enum.OrderKind, e price.Edition) string {
	var prefix string

	switch k {
	case enum.OrderKindCreate:
	case enum.OrderKindRenew:
	case enum.OrderKindUpgrade:
		prefix = "订阅"

	case enum.OrderKindAddOn:
		prefix = "购买"

	default:
	}

	return fmt.Sprintf("%sFT中文网%s", prefix, e.StringCN())
}

// Checkout contains the calculation result of a purchase transaction.
type Checkout struct {
	Kind     enum.OrderKind `json:"kind"`
	Cart     cart.Cart      `json:"cart"`
	Payable  price.Charge   `json:"payable"`
	LiveMode bool           `json:"live"`
}

func (c Checkout) WithTest(t bool) Checkout {
	c.LiveMode = !t

	if t {
		c.Payable.Amount = 0.01
	}

	return c
}
