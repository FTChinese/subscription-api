package input

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
)

type MemberParams struct {
	Tier       enum.Tier      `json:"tier"`
	Cycle      enum.Cycle     `json:"cycle"`
	ExpireDate chrono.Date    `json:"expireDate"`
	PayMethod  enum.PayMethod `json:"payMethod"`
	PriceID    string         `json:"priceId"`   // TODO: send by client
	CreatedBy  string         `json:"createdBy"` // Only exists when membership deleted/updated by ftc staff.
}

func (i MemberParams) Validate() *render.ValidationError {

	if i.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "Tier is required",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	if i.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Cycle is required",
			Field:   "cycle",
			Code:    render.CodeMissingField,
		}
	}

	if i.Tier == enum.TierPremium && i.Cycle == enum.CycleMonth {
		return &render.ValidationError{
			Message: "Premium edition does not have monthly billing cycle",
			Field:   "cycle",
			Code:    render.CodeInvalid,
		}
	}

	if i.PayMethod != enum.PayMethodAli && i.PayMethod != enum.PayMethodWx {
		return &render.ValidationError{
			Message: "Payment method must be one of alipay or wechat",
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}
