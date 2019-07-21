package paywall

// ConfirmationLetter is the content of the email send to user when user successfully subscribed to membership.
const letterNewSub = `
FT中文网用户 {{.User.NormalizeName}},

感谢您订阅FT中文网会员服务。

您于 {{.Sub.CreatedAt.StringCN}} 通过 {{.Sub.PaymentMethod.StringCN}} 订阅了FT中文网 {{.Plan.Desc}}。

订单号 {{.Sub.ID}}
支付金额 {{.Sub.ReadableAmount}}
订阅周期: {{.Sub.StartDate}} 至 {{.Sub.EndDate}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网`

const letterRenewalSub = `
FT中文网用户 {{.User.NormalizeName}},

感谢您续订FT中文网会员服务。

您于 {{.Sub.CreatedAt.StringCN}} 通过 {{.Sub.PaymentMethod.StringCN}} 续订了FT中文网 {{.Plan.Desc}}。

订单号 {{.Sub.ID}}
支付金额 {{.Sub.ReadableAmount}}
订阅周期: {{.Sub.StartDate}} 至 {{.Sub.EndDate}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的持续支持。

FT中文网`

const letterUpgradeSub = `
FT中文网用户 {{.User.NormalizeName}},

感谢您升级订阅FT中文网高端会员。

您于 {{.Sub.CreatedAt.StringCN}} 通过 {{.Sub.PaymentMethod.StringCN}} 从标准会员升级到 {{.Plan.Desc}}。

订单号 {{.Sub.ID}}
支付金额 {{.Sub.ReadableAmount}}
订阅周期: {{.Sub.StartDate}} 至 {{.Sub.EndDate}}

本次升级前余额 {{.Upgrade.ReadableBalance}}，余额来自如下订单未使用部分：

{{.Upgrade.SourceOrderIDs}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的持续支持。

FT中文网`

// Data used to compile this template:
// FtcUser to get user name;
// StripeSub to get period start and end
// stripe.Invoice to get price
const letterStripeSub = `
FT中文网用户 {{.User.NormalizeName}},

您使用Stripe订阅了FT中文网的会员服务，感谢您的支持。

本次订阅创建于 {{.Sub.Created.StringCN}}

订阅产品 {{.Plan.Desc}}
自动续订 {{if .Sub.CancelAtPeriodEnd}}未开启{{else}}已开启{{end}}
订阅周期 {{.Sub.CurrentPeriodStart.StringCN}} - {{.Sub.CurrentPeriodEnd.StringCN}}
订阅状态 {{.Sub.ReadableStatus}}

{{if .Sub.RequiresAction -}}
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
订阅产品 {{.Plan.Desc}}
发票状态 {{.Invoice.Status}}
支付金额 {{.Invoice.Price}}
发票链接 {{.Invoice.HostedInvoiceURL}}
下载PDF {{.Invoice.InvoicePDF}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网`
