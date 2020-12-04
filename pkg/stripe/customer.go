package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/validator"
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
