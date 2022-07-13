package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

const colDiscount = `
	coupon = :coupon,
	customer_id = :customer_id,
	end_time = :end_time,
	invoice_id = :invoice_id,
	promotion_code_id = :promotion_code_id,
	start_time = :start_time,
	subs_id = :subs_id
`

const StmtUpsertDiscount = `
INSERT INTO premium.stripe_discount
SET id = :id,
` + colDiscount + `
ON DUPLICATE KEY UPDATE
` + colDiscount

const StmtRetrieveDiscount = `
SELECT id,
	coupon,
	customer_id,
	end_time,
	invoice_id,
	promotion_code_id,
	start_time,
	subs_id
FROM premium.stripe_discount
WHERE id = ?
LIMIT 1
`

// Discount represents the actual application of a coupon or promotion code. It contains information about when the discount began, when it will end, and what it is applied to.
type Discount struct {
	IsFromStripe    bool         `json:"-"`
	ID              string       `json:"id" db:"id"`
	Coupon          CouponColumn `json:"coupon" db:"coupon"`
	CustomerID      string       `json:"customerId" db:"customer_id"`
	End             null.Int     `json:"end" db:"end_time"`
	InvoiceID       null.String  `json:"invoiceId" db:"invoice_id"`
	InvoiceItemID   null.String  `json:"invoiceItemId" db:"invoice_item_id"`
	PromotionCodeID null.String  `json:"promotionCodeId" db:"promotion_code_id"`
	Start           int64        `json:"start" db:"start_time"`
	SubsID          null.String  `json:"subsId" db:"subs_id"`
}

func NewDiscount(d *stripe.Discount) Discount {
	if d == nil {
		return Discount{}
	}

	var pcID string
	if d.PromotionCode != nil {
		pcID = d.PromotionCode.ID
	}

	return Discount{
		ID:              d.ID,
		Coupon:          CouponColumn{price.NewStripeCoupon(d.Coupon)},
		CustomerID:      d.Customer,
		End:             null.NewInt(d.End, d.End != 0),
		InvoiceID:       null.NewString(d.Invoice, d.Invoice != ""),
		InvoiceItemID:   null.NewString(d.InvoiceItem, d.InvoiceItem != ""),
		PromotionCodeID: null.NewString(pcID, pcID != ""),
		Start:           d.Start,
		SubsID:          null.NewString(d.Subscription, d.Subscription != ""),
	}
}

func (d Discount) IsZero() bool {
	return d.ID != ""
}
