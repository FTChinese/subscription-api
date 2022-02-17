package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"strings"
)

// DefaultPaymentMethodParams contains a customer's default payment method.
type DefaultPaymentMethodParams struct {
	DefaultMethod string `json:"defaultPaymentMethod"`
}

func (p *DefaultPaymentMethodParams) Validate() *render.ValidationError {
	p.DefaultMethod = strings.TrimSpace(p.DefaultMethod)

	return validator.New("defaultPaymentMethod").Required().Validate(p.DefaultMethod)
}

// Customer contains the minimal data of a stripe.Customer.
type Customer struct {
	// A flag indicating whether the data is fetching
	// directly from stripe api.
	// It is false if retrieve from our db.
	IsFromStripe           bool            `json:"-"`
	ID                     string          `json:"id" db:"id"`
	FtcID                  string          `json:"ftcId" db:"ftc_user_id"`
	Currency               stripe.Currency `json:"currency" db:"currency"`
	Created                int64           `json:"created" db:"created"`
	DefaultSource          null.String     `json:"defaultSource" db:"default_source_id"`
	DefaultPaymentMethodID null.String     `json:"defaultPaymentMethod" db:"default_payment_method_id"` // TODO: JSON tag will be renamed to defaultPaymentMethodId in v7
	Email                  string          `json:"email" db:"email"`
	LiveMode               bool            `json:"liveMode" db:"live_mode"`
}

func NewCustomer(ftcID string, c *stripe.Customer) Customer {
	var srcID string
	if c.DefaultSource != nil {
		srcID = c.DefaultSource.ID
	}

	var pmID string
	if c.InvoiceSettings != nil && c.InvoiceSettings.DefaultPaymentMethod != nil {
		pmID = c.InvoiceSettings.DefaultPaymentMethod.ID
	}

	return Customer{
		IsFromStripe:           true,
		ID:                     c.ID,
		FtcID:                  ftcID,
		Currency:               c.Currency,
		Created:                c.Created,
		DefaultSource:          null.NewString(srcID, srcID != ""),
		DefaultPaymentMethodID: null.NewString(pmID, pmID != ""),
		Email:                  c.Email,
		LiveMode:               c.Livemode,
	}
}

// WithFtcID add ftc id to a customer if it does not exist yet.
// Used when you fetched customer from Stripe API but does not
// get ftc id yet.
func (c Customer) WithFtcID(id string) Customer {
	if c.FtcID != "" {
		return c
	}

	c.FtcID = id

	return c
}
