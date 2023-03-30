package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
)

// CouponRedeemed records which invoice has redeemed
// a coupon.
// Once an invoice has any coupon applied to it, it can
// no longer use any others.
type CouponRedeemed struct {
	FtcID       string      `json:"ftcId" db:"ftc_user_id"`
	InvoiceID   string      `json:"invoiceId" db:"invoice_id"`
	LiveMode    bool        `json:"liveMode" db:"live_mode"`
	SubsID      string      `json:"subsId" db:"subs_id"`
	CouponID    string      `json:"couponId" db:"coupon_id"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	RedeemedUTC chrono.Time `json:"redeemedUtc" db:"redeemed_utc"`
}

func (c CouponRedeemed) IsZero() bool {
	return c.InvoiceID == ""
}

const StmtInsertCouponRedeemed = `
INSERT INTO premium.stripe_coupon_redeemed
SET ftc_user_id = :ftc_user_id,
	invoice_id = :invoice_id,
	live_mode = :live_mode,
	subs_id = :subs_id,
	coupon_id = :coupon_id,
	created_utc = :created_utc,
	redeemed_utc = :redeemed_utc
`

const StmtLatestCouponRedeemed = `
SELECT ftc_user_id,
	invoice_id,
	live_mode,
	subs_id,
	coupon_id,
	created_utc,
	redeemed_utc
FROM premium.stripe_coupon_redeemed
WHERE invoice_id = ?
ORDER BY redeemed_utc DESC
LIMIT 1
`
