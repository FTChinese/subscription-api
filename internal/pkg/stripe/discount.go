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

type Discount struct {
	IsFromStripe    bool         `json:"-"`
	ID              string       `db:"id"`
	Coupon          CouponColumn `db:"coupon"`
	CustomerID      string       `db:"customer_id"`
	End             null.Int     `db:"end_time"`
	InvoiceID       null.String  `db:"invoice_id"`
	InvoiceItemID   null.String  `db:"invoice_item_id"`
	PromotionCodeID null.String  `db:"promotion_code_id"`
	Start           int64        `db:"start_time"`
	SubsID          null.String  `db:"subs_id"`
}

func NewDiscount(d *stripe.Discount) Discount {
	return Discount{
		ID:              d.ID,
		Coupon:          CouponColumn{price.NewStripeCoupon(d.Coupon)},
		CustomerID:      d.Customer,
		End:             null.NewInt(d.End, d.End != 0),
		InvoiceID:       null.NewString(d.Invoice, d.Invoice != ""),
		InvoiceItemID:   null.NewString(d.InvoiceItem, d.InvoiceItem != ""),
		PromotionCodeID: null.NewString(d.PromotionCode.ID, d.PromotionCode.ID != ""),
		Start:           d.Start,
		SubsID:          null.NewString(d.Subscription, d.Subscription != ""),
	}
}
