package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strings"
)

// DefaultPaymentMethodParams contains a customer's default payment method.
type DefaultPaymentMethodParams struct {
	CustomerID    string
	DefaultMethod string `json:"defaultPaymentMethod"`
}

func (p *DefaultPaymentMethodParams) Validate() *render.ValidationError {
	p.CustomerID = strings.TrimSpace(p.CustomerID)
	p.DefaultMethod = strings.TrimSpace(p.DefaultMethod)

	ve := validator.New("customerId").Required().Validate(p.CustomerID)
	if ve != nil {
		return ve
	}

	return validator.New("defaultPaymentMethod").Required().Validate(p.DefaultMethod)
}

// Customer contains the minimal data of a stripe.Customer.
type Customer struct {
	ID                   string          `json:"id"`
	FtcID                string          `json:"ftcId"`
	Currency             stripe.Currency `json:"currency"`
	Created              int64           `json:"created"`
	DefaultSource        null.String     `json:"defaultSource"`
	DefaultPaymentMethod null.String     `json:"defaultPaymentMethod"`
	Email                string          `json:"email"`
	LiveMode             bool            `json:"liveMode"`
	CreatedUTC           chrono.Time     `json:"createdUtc"` // Deprecated
}

func NewCustomer(a account.BaseAccount, c *stripe.Customer) Customer {
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
		Currency:             c.Currency,
		Created:              c.Created,
		DefaultSource:        null.NewString(srcID, srcID != ""),
		DefaultPaymentMethod: null.NewString(pmID, pmID != ""),
		Email:                a.Email,
		LiveMode:             c.Livemode,
		CreatedUTC:           chrono.TimeFrom(dt.FromUnix(c.Created)),
	}
}

type CustomerAccount struct {
	BaseAccount account.BaseAccount
	Customer    Customer
}

func NewCustomerAccount(a account.BaseAccount, c *stripe.Customer) CustomerAccount {
	a.StripeID = null.StringFrom(c.ID)

	return CustomerAccount{
		BaseAccount: a,
		Customer:    NewCustomer(a, c),
	}
}
