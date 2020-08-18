package stripe

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/stripe/stripe-go"
	"strings"
)

type Invoice struct {
	*stripe.Invoice
}

func (i Invoice) CreationTime() chrono.Time {
	return chrono.TimeFrom(CanonicalizeUnix(i.Created))
}

func (i Invoice) BuildFtcPlan() (plan.Plan, error) {
	if i.Lines == nil {
		return plan.Plan{}, errors.New("empty lines")
	}

	if len(i.Lines.Data) == 0 {
		return plan.Plan{}, errors.New("empty lines.data")
	}

	stripePlan := i.Lines.Data[0].Plan

	ftcPlan, err := plan.FindPlanForStripe(stripePlan.ID, i.Livemode)

	if err != nil {
		return plan.Plan{}, err
	}

	return ftcPlan.WithStripePrice(*stripePlan), nil
}

func (i Invoice) Price() string {
	return fmt.Sprintf("%s%.2f", strings.ToUpper(string(i.Currency)), float64(i.Total/100))
}

// ReadableStatus turns stripe invoice status into human readable
// text.
// See https://stripe.com/docs/billing/invoices/workflow#workflow-overview
func (i Invoice) ReadableStatus() string {
	switch i.Status {
	case stripe.InvoiceBillingStatusDraft:
		return "草稿"

	case stripe.InvoiceBillingStatusOpen:
		return "等待支付"

	case stripe.InvoiceBillingStatusPaid:
		return "已支付"

	case stripe.InvoiceBillingStatusUncollectible:
		return "无法收款"

	case stripe.InvoiceBillingStatusVoid:
		return "撤销"
	}

	return "未知"
}
