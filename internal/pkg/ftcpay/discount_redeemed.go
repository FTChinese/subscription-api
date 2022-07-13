package ftcpay

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/price"
)

const StmtInsertDiscountRedeemed = `
INSERT INTO premium.ftc_discount_redeemed
SET compound_id = :compound_id,
	discount_id = :discount_id,
	order_id = :order_id,
	redeemed_utc = :redeemed_utc
`

const StmtRetrieveDiscountRedeemed = `
SELECT compound_id,
	discount_id,
	order_id,
	redeemed_utc
FROM premium.ftc_discount_redeemed
WHERE FIND_IN_SET(compound_id, ?)
	AND discount_id = ?
LIMIT 1
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

func NewDiscountRedeemed(order Order, discount price.Discount) DiscountRedeemed {
	return DiscountRedeemed{
		CompoundID:  order.GetCompoundID(),
		DiscountID:  discount.ID,
		OrderID:     order.ID,
		RedeemedUTC: chrono.TimeNow(),
	}
}
