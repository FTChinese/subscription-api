//go:build !production

package stripe

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

func MockRandomDiscount() Discount {
	return Discount{
		IsFromStripe:    false,
		ID:              faker.StripeDiscountID(),
		Coupon:          CouponColumn{price.MockRandomStripeCoupon()},
		CustomerID:      faker.StripeCustomerID(),
		End:             null.Int{},
		InvoiceID:       null.String{},
		InvoiceItemID:   null.String{},
		PromotionCodeID: null.String{},
		Start:           0,
		SubsID:          null.String{},
	}
}
