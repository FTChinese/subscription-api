package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

type Subscription struct {
	Environment           Environment `db:"environment"`
	OriginalTransactionID string      `db:"original_transaction_id"`
	LastTransactionID     string      `db:"last_transaction_id"`
	ProductID             string      `db:"product_id"`
	PurchaseDateUTC       chrono.Time `db:"purchase_date_utc"`
	ExpiresDateUTC        chrono.Time `db:"expires_date_utc"`
	//FtcID                 null.String `db:"ftc_id"`
	//UnionID               null.String `db:"union_id"`
	reader.MemberID
	Tier        enum.Tier  `db:"tier"`
	Cycle       enum.Cycle `db:"cycle"`
	AutoRenewal null.Bool  `db:"auto_renewal"`
}

// Membership build ftc's membership based on subscription
// from apple.
func (s Subscription) Membership() paywall.Membership {
	m := paywall.Membership{
		ID:           null.StringFrom(paywall.GenerateMembershipIndex()),
		MemberID:     s.MemberID,
		LegacyTier:   null.Int{},
		LegacyExpire: null.Int{},
		BasePlan: plan.BasePlan{
			Tier:  s.Tier,
			Cycle: s.Cycle,
		},
		ExpireDate:    chrono.DateFrom(s.ExpiresDateUTC.Time),
		PaymentMethod: enum.PayMethodApple,
		StripeSubID:   null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   s.AutoRenewal.ValueOrZero(),
		Status:        paywall.SubStatusNull,
	}

	m.Normalize()

	return m
}
