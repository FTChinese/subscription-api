package ftcpay

import "github.com/FTChinese/go-rest/chrono"

const StmtInsertDiscountRedeemed = `
INSERT INTO premium.ftc_discount_redeemed
SET compound_id = :compound_id,
	discount_id = :discount_id,
	order_id = :order_id,
	redeemed_utc = :redeemed_utc
`

const StmtDiscountRedeemed = `
SELECT EXISTS (
	SELECT *
	FROM premium.ftc_discount_redeemed
	WHERE discount_id = ?
		AND FIND_IN_SET(compound_id, ?)
)
`

// DiscountRedeemed records user's redemption history of
// discount. During the lifetime of a user, a discount
// could only be redeemed exactly once.
// CompoundID and DiscountID are uniquely constrained.
type DiscountRedeemed struct {
	CompoundID  string      `json:"compoundId" db:"compound_id"`
	DiscountID  string      `json:"discountId" db:"discount_id"`
	OrderID     string      `json:"orderId" db:"order_id"`
	RedeemedUTC chrono.Time `json:"redeemedUtc" db:"redeemed_utc"`
}
