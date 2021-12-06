package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/stripe/stripe-go/v72"
)

type CheckoutSession struct {
	CancelURL     string                              `json:"cancelUrl"`
	AmountTotal   int64                               `json:"amountTotal"`
	Currency      price.Currency                      `json:"currency"`
	CustomerID    string                              `json:"customerId"`
	ID            string                              `json:"id"`
	LiveMode      bool                                `json:"liveMode"`
	Mode          stripe.CheckoutSessionMode          `json:"mode"`
	PaymentStatus stripe.CheckoutSessionPaymentStatus `json:"paymentStatus"`
	SuccessURL    string                              `json:"successUrl"`
}

func NewCheckoutSession(sc *stripe.CheckoutSession) CheckoutSession {
	return CheckoutSession{
		CancelURL:     sc.CancelURL,
		AmountTotal:   sc.AmountTotal,
		Currency:      price.Currency(sc.Currency),
		CustomerID:    sc.Customer.ID,
		ID:            sc.ID,
		LiveMode:      false,
		Mode:          sc.Mode,
		PaymentStatus: sc.PaymentStatus,
		SuccessURL:    sc.SuccessURL,
	}
}
