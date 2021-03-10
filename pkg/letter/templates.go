package letter

var templates = map[string]string{
	keyNewSubs: `
FT中文网用户 {{.UserName}},

感谢您订阅FT中文网会员服务。

您于 {{.Purchased.CreatedUTC.StringCN}} 通过 {{.Purchased.PaymentMethod.StringCN}} 订阅了FT中文网 {{.Purchased.Tier.StringCN}}。

订单号 {{.Order.ID}}
支付金额 {{.Order.Amount | currency}}
订阅周期: {{.Order.StartDate}} 至 {{.Order.EndDate}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网`,
	keyRenewalSubs: `
FT中文网用户 {{.UserName}},

感谢您续订FT中文网会员服务。

您于 {{.Purchased.CreatedUTC.StringCN}} 通过 {{.Purchased.PaymentMethod.StringCN}} 续订了FT中文网 {{.Purchased.Tier.StringCN}}。

订单号 {{.Order.ID}}
支付金额 {{.Order.Amount | currency}}
订阅周期: {{.Order.StartDate}} 至 {{.Order.EndDate}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的持续支持。

FT中文网`,
	keyUpgradeSubs: `
FT中文网用户 {{.UserName}},

感谢您升级订阅FT中文网高端会员。

您于 {{.Order.CreatedAt.StringCN}} 通过 {{.Order.PaymentMethod.StringCN}} 从标准会员升级到 {{.Order.Tier.StringCN}}。

订单号 {{.Order.ID}}
支付金额 {{.Order.Amount | currency}}
订阅周期: {{.Order.StartDate}} 至 {{.Order.EndDate}}

本次升级前标准版订阅剩余 {{.CarriedOver.TotalDays}} 天，将在高端版到期后再次启用

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的持续支持。

FT中文网`,
	keyAddOn: `
FT中文网用户 {{.UserName}},

感谢您购买FT中文网会员服务。

您于 {{.Purchased.CreatedUTC.StringCN}} 通过 {{.Purchased.PaymentMethod.StringCN}} 购买一份 {{.Purchased.Tier.StringCN}}。

订单号 {{.Order.ID}}
支付金额 {{.Order.Amount | currency}}
购买天数: {{.Purchased.TotalDays}}

您当前会员失效后将启用本次购买的 {{.Purchased.Tier.StringCN}}。

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的持续支持。

FT中文网
`,
	keyIAPLinked: `
FT中文网用户 {{.UserName}},

您的FT中文网账号 {{.Email}} 已经关联了在iOS平台上订阅的FT中文网会员服务。

订阅产品：{{.Tier.StringCN}}
到期日期：{{.ExpireDate}}

在其他平台使用FT中文网账号登录即可实现跨平台阅读。

感谢您对FT中文网的支持。如需帮助，请联系客服：subscriber.service@ftchinese.com。`,
	keyIAPUnlinked: `
FT中文网用户 {{.UserName}},

您的FT中文网账号 {{.Email}} 已经移除了一项关联的iOS平台的FT中文网订阅。

订阅产品：{{.Tier.StringCN}}
到期日期：{{.ExpireDate}}

您可以在使用该订阅的苹果设备登录FT中文网账号后可以重新绑定。

感谢您对FT中文网的支持。如需帮助，请联系客服：subscriber.service@ftchinese.com。`,
}

// Data used to compile this template:
// Account to get user name;
// StripeSub to get period start and end
// stripe.Invoice to get price
const letterStripeSub = `
FT中文网用户 {{.User.NormalizeName}},

您使用Stripe订阅了FT中文网的会员服务，感谢您的支持。

本次订阅创建于 {{.Order.Created.StringCN}}

订阅产品 {{.Price.Desc}}
自动续订 {{if .Order.CancelAtPeriodEnd}}未开启{{else}}已开启{{end}}
订阅周期 {{.Order.CurrentPeriodStart.StringCN}} - {{.Order.CurrentPeriodEnd.StringCN}}
订阅状态 {{.Order.ReadableStatus}}

{{if .Order.RequiresAction -}}
我们注意到您本次订阅的支付尚未完成，请按照提示进行支付。如果支付遇到问题，可以咨询FT中文网客服。如果您已经完成支付，请忽略。
{{end -}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网`

const letterStripeInvoice = `
FT中文网用户 {{.User.NormalizeName}},

以下是您通过Stripe订阅FT中文网会员的发票信息。

单号 {{.Invoice.Number}}
创建于 {{.Invoice.CreationTime.StringCN}}
订阅产品 {{.Price.Desc}}
发票状态 {{.Invoice.ReadableStatus}}
支付金额 {{.Invoice.Price}}
发票链接 {{.Invoice.HostedInvoiceURL}}
下载PDF {{.Invoice.InvoicePDF}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网`

const letterStripePaymentFailed = `
FT中文网用户 {{.User.NormalizeName}},

您通过Stripe支付订阅FT中文网 {{.Price.Desc}} 支付失败。本次支付的发票号是 {{.Invoice.Number}}，创建于 {{.Invoice.CreationTime.StringCN}}。

您可以联系我们的客服：subscriber.service@ftchinese.com 询问支付问题。

目前FT中文网的Stripe支付以英镑结算，不支持银联(UnionPay)等人民币信用卡。您可以使用有带有Visa、Mastercard、American Express、Discover、Diners Club等标志的卡片。

感谢您对FT中文网的支持。

FT中文网`

const letterPaymentActionRequired = `
FT中文网用户 {{.User.NormalizeName}},

您通过Stripe支付订阅FT中文网 {{.Price.Desc}} 尚未完成支付，您的发卡行可能需要进行安全验证。请按照Stripe的提示执行下一步操作。

本次支付的发票号是 {{.Invoice.Number}}，创建于 {{.Invoice.CreationTime.StringCN}}。

您可以联系我们的客服：subscriber.service@ftchinese.com 询问支付问题。

目前FT中文网的Stripe支付以英镑结算，不支持银联(UnionPay)等人民币信用卡。您可以使用有带有Visa、Mastercard、American Express、Discover、Diners Club等标志的卡片。

感谢您对FT中文网的支持。

FT中文网`
