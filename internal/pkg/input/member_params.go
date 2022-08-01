package input

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
)

// MemberParams is used to create a membership directly.
// This should never be exposed to user.
// FtcID and UnionID exists only upon creation.
type MemberParams struct {
	FtcID      null.String    `json:"ftcId"`
	UnionID    null.String    `json:"unionId"`
	PriceID    string         `json:"priceId"`
	ExpireDate chrono.Date    `json:"expireDate"`
	PayMethod  enum.PayMethod `json:"payMethod"`
}

func (i MemberParams) Validate(isNew bool) *render.ValidationError {

	if isNew && i.FtcID.IsZero() && i.UnionID.IsZero() {
		return &render.ValidationError{
			Message: "Either ftcId or unionId is required",
			Field:   "compoundId",
			Code:    render.CodeMissingField,
		}
	}

	ve := validator.New("priceId").Required().Validate(i.PriceID)
	if ve != nil {
		return ve
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
