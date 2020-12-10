package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strings"
)

// PaymentInput contains a customer's payment method.
type PaymentInput struct {
	CustomerID    string
	DefaultMethod string `json:"defaultPaymentMethod"`
}

func (p *PaymentInput) Validate() *render.ValidationError {
	p.CustomerID = strings.TrimSpace(p.CustomerID)
	p.DefaultMethod = strings.TrimSpace(p.DefaultMethod)

	ve := validator.New("customerId").Required().Validate(p.CustomerID)
	if ve != nil {
		return ve
	}

	return validator.New("defaultPaymentMethod").Required().Validate(p.DefaultMethod)
}

type Customer struct {
	ID                   string      `json:"id"`
	FtcID                string      `json:"ftcId"`
	DefaultSource        null.String `json:"defaultSource"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	Email                string      `json:"email"`
	LiveMode             bool        `json:"liveMode"`
	CreatedUTC           chrono.Time `json:"createdUtc"`
}

func NewCustomer(a reader.FtcAccount, c *stripe.Customer) Customer {
	var srcID string
	if c.DefaultSource != nil {
		srcID = c.DefaultSource.ID
	}

	var pmID string
	if c.InvoiceSettings != nil && c.InvoiceSettings.DefaultPaymentMethod != nil {
		pmID = c.InvoiceSettings.DefaultPaymentMethod.ID
	}

	return Customer{
		ID:                   c.ID,
		FtcID:                a.FtcID,
		DefaultSource:        null.NewString(srcID, srcID != ""),
		DefaultPaymentMethod: null.NewString(pmID, pmID != ""),
		Email:                a.Email,
		LiveMode:             c.Livemode,
		CreatedUTC:           chrono.TimeFrom(dt.FromUnix(c.Created)),
	}
}
