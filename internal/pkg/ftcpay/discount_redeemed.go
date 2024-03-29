package ftcpay

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/price"
)

const StmtInsertDiscountRedeemed = `
INSERT INTO premium.ftc_discount_redeemed
SET compound_id = :compound_id,
	discount_id = :discount_id,
	live_mode = :live_mode,
	order_id = :order_id,
	redeemed_utc = :redeemed_utc
`

const StmtRetrieveDiscountRedeemed = `
SELECT compound_id,
	discount_id,
	live_mode,
	order_id,
	redeemed_utc
FROM premium.ftc_discount_redeemed
WHERE FIND_IN_SET(compound_id, ?)
	AND discount_id = ?
LIMIT 1
`

// DiscountRedeemed records user's redemption history of
// non-recurring discount.
// During the lifetime of a user, a non-recurring discount
// could only be redeemed exactly once.
// CompoundID and DiscountID are uniquely constrained.
type DiscountRedeemed struct {
	CompoundID  string      `json:"compoundId" db:"compound_id"`
	DiscountID  string      `json:"discountId" db:"discount_id"`
	LiveMode    bool        `json:"liveMode" db:"live_mode"`
	OrderID     string      `json:"orderId" db:"order_id"`
	RedeemedUTC chrono.Time `json:"redeemedUtc" db:"redeemed_utc"`
}

func NewDiscountRedeemed(order Order, discount price.Discount) DiscountRedeemed {
	if discount.IsZero() || discount.Recurring {
		return DiscountRedeemed{}
	}

	return DiscountRedeemed{
		CompoundID:  order.GetCompoundID(),
		DiscountID:  discount.ID,
		LiveMode:    discount.LiveMode,
		OrderID:     order.ID,
		RedeemedUTC: chrono.TimeNow(),
	}
}

func (dr DiscountRedeemed) IsZero() bool {
	return dr.DiscountID == ""
}
