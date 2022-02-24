package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg"
	"github.com/FTChinese/subscription-api/lib/validator"
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

type PaymentMethodCard struct {
	Brand             stripe.PaymentMethodCardBrand             `json:"brand"`
	Country           string                                    `json:"country"`
	ExpMonth          uint64                                    `json:"expMonth"`
	ExpYear           uint64                                    `json:"expYear"`
	Fingerprint       string                                    `json:"fingerprint"`
	Funding           stripe.CardFunding                        `json:"funding"`
	Last4             string                                    `json:"last4"`
	Network           stripe.PaymentMethodCardNetworks          `json:"network"`
	ThreeDSecureUsage stripe.PaymentMethodCardThreeDSecureUsage `json:"threeDSecureUsage"`
}

func NewPaymentMethodCard(c *stripe.PaymentMethodCard) PaymentMethodCard {
	if c == nil {
		return PaymentMethodCard{}
	}

	var network stripe.PaymentMethodCardNetworks
	if c.Networks != nil {
		network = *c.Networks
	}

	var threeD stripe.PaymentMethodCardThreeDSecureUsage
	if c.ThreeDSecureUsage != nil {
		threeD = *c.ThreeDSecureUsage
	}

	return PaymentMethodCard{
		Brand:             c.Brand,
		Country:           c.Country,
		ExpMonth:          c.ExpMonth,
		ExpYear:           c.ExpYear,
		Fingerprint:       c.Fingerprint,
		Funding:           c.Funding,
		Last4:             c.Last4,
		Network:           network,
		ThreeDSecureUsage: threeD,
	}
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (c PaymentMethodCard) Value() (driver.Value, error) {

	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (c *PaymentMethodCard) Scan(src interface{}) error {
	if src == nil {
		*c = PaymentMethodCard{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp PaymentMethodCard
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*c = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentMethodCard")
	}
}

// PaymentMethod is a reduced version of stripe payment method.
// You need to find a payment method's id to get it.
// The payment method is hides in various places:
// - Customer.invoice_settings.default_payment_method
// - PaymentIntent.payment_method
// - Subscription.default_payment_method
// When you created a new subscription, you have to get it
// by expanding latest_invoice.payment_intent field to get
// the id.
type PaymentMethod struct {
	IsFromStripe bool                     `json:"-"` // A flag indicating the data comes from Stripe API.
	ID           string                   `json:"id" db:"id"`
	CustomerID   string                   `json:"customerId" db:"customer_id"`
	Kind         stripe.PaymentMethodType `json:"kind" db:"kind"`
	Card         PaymentMethodCard        `json:"card" db:"card_details"`
	Created      int64                    `json:"created" db:"created"`
	LiveMode     bool                     `json:"liveMode" db:"live_mode"`
}

func NewPaymentMethod(pm *stripe.PaymentMethod) PaymentMethod {
	if pm == nil {
		return PaymentMethod{}
	}

	if pm.ID == "" {
		return PaymentMethod{}
	}

	var cusID string
	if pm.Customer != nil {
		cusID = pm.Customer.ID
	}

	return PaymentMethod{
		IsFromStripe: true,
		ID:           pm.ID,
		CustomerID:   cusID,
		Kind:         pm.Type,
		Card:         NewPaymentMethodCard(pm.Card),
		Created:      pm.Created,
		LiveMode:     false,
	}
}

func (pm PaymentMethod) IsZero() bool {
	return pm.ID == ""
}

// PaymentMethodJSON is used to save PaymentMethod as JSON into a single column.
// It exists due to PaymentMethod is used to save as a normal sql row.
type PaymentMethodJSON struct {
	PaymentMethod
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (pm PaymentMethodJSON) Value() (driver.Value, error) {
	if pm.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(pm.PaymentMethod)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (pm *PaymentMethodJSON) Scan(src interface{}) error {
	if src == nil {
		*pm = PaymentMethodJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp PaymentMethod
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*pm = PaymentMethodJSON{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentMethodCard")
	}
}

type PagedPaymentMethods struct {
	pkg.PagedList
	Data []PaymentMethod `json:"data"`
}
