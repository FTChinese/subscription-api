package model

const letter = `
Dear FTC user {{.Name}},

{{if .IsReneal}}
You have renewed your subcription to FTC membership.
{{- else}}
You have subscriped to FTC membership.
{{- end}} Thanks.

Here is your order details:

Order ID: {{.OrderID}}
Membership: {{.MemberTier}}/{{.BillingCycle}}
Price: {{.TotalAmount}}
Payment Method: {{.PaymentMethod}}
Duration: {{.StartDate}} to {{.EndDate}}

If you have any questions, please contact our customer service:

Wechat: xxxx
Phone: xxxxx

Againt, we appriciate your support.
`
