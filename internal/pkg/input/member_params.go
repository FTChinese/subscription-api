package input

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
)

// MemberParams is used to create a membership directly.
// This should never be exposed to user.
// FtcID and UnionID exists only upon creation.
type MemberParams struct {
	FtcID      null.String    `json:"ftcId"`
	UnionID    null.String    `json:"unionId"`
	Tier       enum.Tier      `json:"tier"`
	Cycle      enum.Cycle     `json:"cycle"`
	ExpireDate chrono.Date    `json:"expireDate"`
	PayMethod  enum.PayMethod `json:"payMethod"`
	PriceID    string         `json:"-"`
}

func (i MemberParams) Validate(isNew bool) *render.ValidationError {

	if isNew && i.FtcID.IsZero() && i.UnionID.IsZero() {
		return &render.ValidationError{
			Message: "Either ftcId or unionId is required",
			Field:   "compoundId",
			Code:    render.CodeMissingField,
		}
	}

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
