package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"time"
)

type StripeCouponMeta struct {
	PriceID  null.String `json:"priceId" db:"price_id"`
	StartUTC chrono.Time `json:"startUtc" db:"start_utc"`
}

func ParseStripeCouponMeta(m map[string]string) StripeCouponMeta {
	priceId := m["price_id"]
	start := m["start_utc"]

	var startTime time.Time
	if start != "" {
		startTime, _ = time.Parse(time.RFC3339, start)
	}

	return StripeCouponMeta{
		PriceID:  null.NewString(priceId, priceId != ""),
		StartUTC: chrono.TimeFrom(startTime),
	}
}

func (p StripeCouponMeta) ToMap() map[string]string {
	return map[string]string{
		"price_id":  p.PriceID.String,
		"start_utc": p.StartUTC.Format(time.RFC3339),
	}
}

// StripeCoupon reduces the amount charged to a customer by discounting their subscription.
// https://stripe.com/docs/billing/subscriptions/coupons
// A coupon is equivalent to ftc pricing's discount
// targeting retention.
// We should limit multiple redemption of coupons. By multiple redemption we mean two cases:
// 1. A subscription redeemed the same coupons more than once;
// 2. A subscription, in a single billing cycle (one year or one month), redeemed multiple coupons.
//
// To solve the first problem, a coupon should be redeemed once and only once
// during the lifecycle of a stripe subscription, which could be imposed by recording
// a one-to-one mapping between a subscription id to a coupon id.
// However, this gives rise to problem 2:
// suppose a user's current billing cycle starts from Jan to Dec, and you issued one coupon each month,
// then this user could redeem 12 coupons.
// To solve problem 2, a mapping between subscription id to coupons id is not enough.
// It might be reasonable if we could establish a mapping between a specific billing cycle and a coupon,
// so that a billing cycle could not redeem another coupon as long as one is already applied to it.
type StripeCoupon struct {
	IsFromStripe bool        `json:"-"`
	ID           string      `json:"id" db:"id"`
	AmountOff    int64       `json:"amountOff" db:"amount_off"`
	Created      int64       `json:"created" db:"created"`
	Currency     string      `json:"currency" db:"currency"`
	Duration     string      `json:"duration" db:"duration"`
	EndUTC       chrono.Time `json:"endUtc" db:"endUtc"` // Stripe's redeem_by field.
	LiveMode     bool        `json:"liveMode" db:"live_mode"`
	Name         string      `json:"name" db:"name"`
	StripeCouponMeta
	Status DiscountStatus `json:"status" db:"status"`
}

func NewStripeCoupon(c *stripe.Coupon) StripeCoupon {
	meta := ParseStripeCouponMeta(c.Metadata)

	status := DiscountStatusActive
	if !c.Valid || c.Deleted {
		status = DiscountStatusCancelled
	}

	return StripeCoupon{
		IsFromStripe:     true,
		ID:               c.ID,
		AmountOff:        c.AmountOff,
		Created:          c.Created,
		Currency:         string(c.Currency),
		Duration:         string(c.Duration),
		EndUTC:           chrono.TimeFrom(time.Unix(c.RedeemBy, 0)),
		LiveMode:         c.Livemode,
		Name:             c.Name,
		StripeCouponMeta: meta,
		Status:           status,
	}
}

func (c StripeCoupon) IsZero() bool {
	return c.ID == ""
}

func (c StripeCoupon) Cancelled() StripeCoupon {
	c.Status = DiscountStatusCancelled
	return c
}
