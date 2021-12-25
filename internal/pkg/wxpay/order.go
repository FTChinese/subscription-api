package wxpay

import (
	"github.com/go-pay/gopay"
)

// OrderParamsCommon collects the essential fields to create
// a new order for ftc.
type OrderParamsCommon struct {
	Body        string `map:"body"`             // 商品描述交易字段格式根据不同的应用场景按照以下格式： APP——需传入应用市场上的APP名字-实际商品名称，天天爱消除-游戏充值。
	OutTradeNo  string `map:"out_trade_no"`     // 商户系统内部订单号，要求32个字符内（最少6个字符），只能是数字、大小写字母_-|*且在同一个商户号下唯一
	TotalAmount int64  `map:"total_fee"`        // 订单总金额，单位为分
	IP          string `map:"spbill_create_ip"` // 支持IPV4和IPV6两种格式的IP地址。调用微信支付API的机器IP
	CallbackURL string `map:"notify_url"`       // 接收微信支付异步通知回调地址，通知url必须为直接可访问的url，不能携带参数。公网域名必须为https
	TradeType   string `map:"trade_type"`
}

// OrderParams lists all request parameters sent to
// https://api.mch.weixin.qq.com/pay/unifiedorder
type OrderParams struct {
	OrderParamsCommon
	AppID       string `map:"appid"`                    // 微信开放平台审核通过的应用APPID
	MchID       string `map:"mch_id"`                   // 微信支付分配的商户号
	DeviceInfo  string `map:"device_info,omitempty"`    // optional 终端设备号(门店号或收银设备ID)，默认请传"WEB"
	Nonce       string `map:"nonce_str"`                // 随机字符串，不长于32位
	Sign        string `map:"-"`                        // sign, 签名
	SignType    string `map:"sign_type"`                // 签名类型，目前支持HMAC-SHA256和MD5
	Attach      string `map:"attach,omitempty"`         // 附加数据，在查询API和支付通知中原样返回，该字段主要用于商户携带订单的自定义数据
	Currency    string `map:"fee_type,omitempty"`       // 符合ISO 4217标准的三位字母代码，默认人民币：CNY
	StartTime   string `map:"time_start,omitempty"`     // 订单生成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
	EndTime     string `map:"time_expire,omitempty"`    // 订单失效时间
	OfferTag    string `map:"goods_tag,omitempty"`      // 订单优惠标记，代金券或立减优惠功能的参数
	LimitPay    string `map:"limit_pay,omitempty"`      // no_credit--指定不能使用信用卡支付
	ShowReceipt bool   `map:"receipt,omitempty"`        // 开发票入口开放标识. Y，传入Y时，支付成功消息和支付详情页将出现开票入口。需要在微信支付商户平台或微信公众平台开通电子发票功能，传此字段才可生效
	ShareProfit bool   `map:"profit_sharing,omitempty"` // Y-是，需要分账, N-否，不分账
}

func NewOrderParams(o OrderParamsCommon, app AppConfig) OrderParams {
	return OrderParams{
		OrderParamsCommon: o,
		AppID:             app.AppID,
		MchID:             app.MchID,
		Nonce:             NonceStr(),
		Sign:              "",
		SignType:          app.SignType,
	}
}

func (o OrderParams) Marshal() gopay.BodyMap {
	return Marshal(o)
}
