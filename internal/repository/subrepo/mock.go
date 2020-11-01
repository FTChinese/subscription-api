// +build !production

package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
)

func (c WxPayClient) MockWebhookPayload(n wechat.Notification) wxpay.Params {
	p := n.MockToMap()
	s := c.sdk.Sign(p)

	p.SetString("sign", s)

	return p
}

func (c WxPayClient) MockOrderPayload(or wechat.OrderResp) wxpay.Params {
	p := or.MockToMap()

	s := c.sdk.Sign(p)
	p.SetString("sign", s)

	return p
}
