package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

type Subscription struct {
	Environment           Environment `db:"environment"`
	OriginalTransactionID string      `db:"original_transaction_id"`
	LastTransactionID     string      `db:"last_transaction_id"`
	ProductID             string      `db:"product_id"`
	PurchaseDateUTC       chrono.Time `db:"purchase_date_utc"`
	ExpiresDateUTC        chrono.Time `db:"expires_date_utc"`
	Tier                  enum.Tier   `db:"tier"`
	Cycle                 enum.Cycle  `db:"cycle"`
	AutoRenewal           bool        `db:"auto_renewal"`
}

// Membership build ftc's membership based on subscription
// from apple.
func (s Subscription) NewMembership(id reader.MemberID) subscription.Membership {
	m := subscription.Membership{
		ID:           null.StringFrom(subscription.GenerateMembershipIndex()),
		LegacyExpire: null.Int{},
		MemberID:     id,
		BasePlan: plan.BasePlan{
			Tier:  s.Tier,
			Cycle: s.Cycle,
		},
		ExpireDate:    chrono.DateFrom(s.ExpiresDateUTC.Time),
		PaymentMethod: enum.PayMethodApple,
		StripeSubID:   null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   s.AutoRenewal,
		Status:        subscription.SubStatusNull,
	}

	m.Normalize()

	return m
}
