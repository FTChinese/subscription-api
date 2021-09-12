package invoice

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

// AddOnParams contains parameters used to create an addon invoice.
// When AddOnSource is addon.SourceUserPurchase, order id should exist;
// WHen AddOnSource is addon.SourceCarryOver or addon.SourceCompensation,
// order id must not exist,
// and Edition, YearMonthDay should be provided.
type AddOnParams struct {
	CompoundID  string       `json:"-"`
	AddOnSource addon.Source `json:"source"`
	price.Edition
	dt.YearMonthDay
	OrderID       null.String    `json:"orderId"`
	PaidAmount    float64        `json:"paidAmount"`
	PaymentMethod enum.PayMethod `json:"payMethod"`
	PriceID       null.String    `json:"priceId"`
}

func (p *AddOnParams) Validate() *render.ValidationError {
	ve := p.validateCommonField()
	if ve != nil {
		return ve
	}

	switch p.AddOnSource {
	case addon.SourceUserPurchase:
		if p.OrderID.IsZero() {
			return &render.ValidationError{
				Message: "Order id is required for user_purchase",
				Field:   "orderId",
				Code:    render.CodeMissingField,
			}
		}

	case addon.SourceCarryOver, addon.SourceCompensation:
		if p.OrderID.Valid {
			return &render.ValidationError{
				Message: "Order id should not exist for carry_over or compensation",
				Field:   "orderId",
				Code:    render.CodeInvalid,
			}
		}
	}

	return nil
}

func (p *AddOnParams) validateCommonField() *render.ValidationError {
	if p.Years <= 0 && p.Months <= 0 && p.Days <= 0 {
		return &render.ValidationError{
			Message: "Must provide one of years, months, days",
			Field:   "years_months_days",
			Code:    render.CodeMissingField,
		}
	}

	if p.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "Tier is required",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	if p.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Cycle is required",
			Field:   "cycle",
			Code:    render.CodeMissingField,
		}
	}

	if p.PaymentMethod == enum.PayMethodNull {
		return &render.ValidationError{
			Message: "Payment method is required",
			Field:   "payMethod",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}
