package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
)

type CancelParams struct {
	FtcID  string
	SubID  string
	Cancel bool // True for cancel, false for reactivation.
}

type CustomerParams struct {
	Customer string `json:"customer"`
}

func (p CustomerParams) Validate() *render.ValidationError {
	return validator.New("customer").Required().Validate(p.Customer)
}
