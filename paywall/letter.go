package paywall

// ConfirmationLetter is the content of the email send to user when user successfully subscribed to membership.
const confirmationLetter = `
FT中文网用户 {{.User.NormalizeName}},

您好！{{if .Subs.IsRenewal -}}
感谢您续订FT中文网会员服务。
{{- else -}}
感谢您订阅FT中文网会员服务。
{{- end}}

您本次订单的详细信息如下：

订单号: {{.Subs.OrderID}}
会员类型: {{.Subs.TierToBuy.ToCN}}/{{.Subs.BillingCycle.ToCN}}
支付金额: ¥{{.Subs.NetPrice}}
支付方式: {{.Subs.PaymentMethod.ToCN}}
订单日期: {{.Subs.CreatedAt.StringCN}}
本次订单购买的会员期限: {{.Subs.StartDate}} 至 {{.Subs.EndDate}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网

---------------------

Dear FTC user {{.User.NormalizeName}},
{{if .Subs.IsRenewal -}}
You have renewed your subcription to FTC membership.
{{- else}}
You have subscriped to FTC membership.
{{- end}} Thanks.

Here is your order details:

Order ID: {{.Subs.OrderID}}
Membership: {{.Subs.TierToBuy.ToEN}}/{{.Subs.BillingCycle.ToEN}}
Price: CNY {{.Subs.TotalAmount}}
Payment Method: {{.Subs.PaymentMethod.ToEN}}
Created At: {{.Subs.CreatedAt.StringEN}}
Duration: {{.Subs.StartDate}} to {{.Subs.EndDate}}

To get help with subscription and purchases, please contact subscriber.service@ftchinese.com.

Again, we appreciate your support.

FTChinese`
