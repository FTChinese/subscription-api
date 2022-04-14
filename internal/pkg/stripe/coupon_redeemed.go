package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
)

// CouponRedeemed records which invoice has redeemed
// a coupon.
// One an invoice has any coupon applied to it, it can
// no longer use any others.
type CouponRedeemed struct {
	InvoiceID   string      `db:"invoice_id"`
	FtcID       string      `db:"ftc_user_id"`
	SubsID      string      `db:"subs_id"`
	CouponID    string      `db:"coupon_id"`
	CreatedUTC  chrono.Time `db:"created_utc"`
	RedeemedUTC chrono.Time `db:"redeemed_utc"`
}

func (c CouponRedeemed) IsZero() bool {
	return c.InvoiceID == ""
}

const StmtInsertCouponRedeemed = `
INSERT INTO premium.stripe_coupon_redeemed
SET invoice_id = :invoice_id,
	ftc_user_id = :ftc_user_id,
	subs_id = :subs_id,
	coupon_id = :coupon_id,
	created_utc = :created_utc,
	redeemed_utc = :redeemed_utc
`

const StmtInvoiceHasCoupon = `
SELECT EXISTS (
	SELECT invoice_id
	FROM premium.stripe_coupon_redeemed
	WHERE invoice_id = ?
) AS alreadyExists
`
