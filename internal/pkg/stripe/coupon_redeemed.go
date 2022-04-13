package stripe

import "github.com/FTChinese/go-rest/chrono"

type CouponRedeemed struct {
	FtcID       string      `db:"ftc_id"`
	SubsID      string      `db:"subs_id"`
	InvoiceID   string      `db:"invoice_id"`
	CouponID    string      `db:"coupon_id"`
	CreatedUTC  chrono.Time `db:"created_utc"`
	RedeemedUTC chrono.Time `db:"redeemed_tc"`
}
