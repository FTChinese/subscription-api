package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/price"
	"strings"
)

type CancelParams struct {
	FtcID  string
	SubID  string
	Cancel bool // True for cancel, false for reactivation.
}

type CheckoutInput struct {
	price.Edition
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

func (i *CheckoutInput) Validate() *render.ValidationError {
	if ve := i.Edition.Validate(); ve != nil {
		return ve
	}

	i.SuccessURL = strings.TrimSpace(i.SuccessURL)
	i.CancelURL = strings.TrimSpace(i.CancelURL)

	ve := validator.New("successUrl").Required().Validate(i.SuccessURL)
	if ve != nil {
		return ve
	}

	ve = validator.New("cancelUrl").Required().Validate(i.CancelURL)
	if ve != nil {
		return ve
	}

	return nil
}

type CheckoutParams struct {
	Account account.BaseAccount
	Plan    PriceEdition
	Input   CheckoutInput
}
