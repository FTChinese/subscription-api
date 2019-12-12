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
	plan.BasePlan
	AutoRenewal bool `db:"auto_renewal"`
}

// Membership build ftc's membership based on subscription
// from apple if this subscription is not linked
// to an ftc account.
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
		AutoRenew:     s.AutoRenewal,
		Status:        subscription.SubStatusNull,
		AppleSubID:    null.StringFrom(s.OriginalTransactionID),
	}

	m.Normalize()

	return m
}

// BuildOn updates an existing IAP membership based on this
// transaction.
func (s Subscription) BuildOn(m subscription.Membership) subscription.Membership {
	m.Tier = s.Tier
	m.Cycle = s.Cycle
	m.ExpireDate = chrono.DateFrom(s.ExpiresDateUTC.Time)
	m.PaymentMethod = enum.PayMethodApple
	m.StripeSubID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenew = s.AutoRenewal
	m.Status = subscription.SubStatusNull
	m.AppleSubID = null.StringFrom(s.OriginalTransactionID)

	return m
}
