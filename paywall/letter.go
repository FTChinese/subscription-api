package paywall

// ConfirmationLetter is the content of the email send to user when user successfully subscribed to membership.
const confirmationLetter = `
FT中文网用户 {{.User.NormalizeName}},

您好！{{if .Subs.IsNewMember -}}
感谢您订阅FT中文网会员服务。
{{- else if .Subs.IsRenewal -}}
感谢您续订FT中文网会员服务。
{{- else if .Subs.IsUpgrade -}}
感谢您升级位高级会员。
{{- end}}

您本次订单的详细信息如下：

订单号: {{.Subs.OrderID}}
会员类型: {{.Subs.TierToBuy.ToCN}}/{{.Subs.BillingCycle.ToCN}}
支付金额: ¥{{.Subs.NetPrice}}
{{if .Subs.IsValidPay -}}
支付方式: {{.Subs.PaymentMethod.ToCN}}
{{- end}}
订单日期: {{.Subs.CreatedAt.StringCN}}
本次订单购买的会员期限: {{.Subs.StartDate}} 至 {{.Subs.EndDate}}

{{if .Subs.IsUpgrade -}}
此前订单未使用部分的余额已折换进本次订单：
订单号: {{.Subs.UpgradeSourceIDs}}
余额共计: ¥{{.Subs.ProratedAmount}}
{{- end}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

再次感谢您对FT中文网的支持。

FT中文网

---------------------

Dear FTC user {{.User.NormalizeName}},
{{if .Subs.IsRenewal -}}
You have renewed your subscription to FTC membership.
{{- else}}
You have subscribed to FTC membership.
{{- end}} Thanks.

Here is your order details:

Order ID: {{.Subs.OrderID}}
Membership: {{.Subs.TierToBuy.ToEN}}/{{.Subs.BillingCycle.ToEN}}
Price: CNY {{.Subs.TotalAmount}}
Payment Method: {{.Subs.PaymentMethod.ToEN}}
Created At: {{.Subs.CreatedAt.StringEN}}
Duration: {{.Subs.StartDate}} to {{.Subs.EndDate}}

To get help with subscription and purchases, please contact subscriber.service@ftchinese.com.

{{if .Subs.IsUpgrade -}}
The unused portion of all your previous orders have been prorated and deducted from this payment：
Those orders are: {{.Subs.UpgradeSourceIDs}}
Total balance: ¥{{.Subs.ProratedAmount}}
{{- end}}

Again, we appreciate your support.

FTChinese`
