package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

// See errors returned from Membership.PermitAliWxUpgrade.
func (env SubEnv) PreviewUpgrade(userID reader.MemberID) (subscription.PaymentIntent, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subscription.PaymentIntent{}, err
	}

	member, err := tx.RetrieveMember(userID)
	if err != nil {
		_ = tx.Rollback()
		return subscription.PaymentIntent{}, err
	}

	if !member.PermitAliWxUpgrade() {
		_ = tx.Rollback()
		return subscription.PaymentIntent{}, subscription.ErrUpgradeInvalid
	}

	orders, err := tx.FindBalanceSources(userID)
	if err != nil {
		_ = tx.Rollback()
		return subscription.PaymentIntent{}, err
	}

	wallet := subscription.NewWallet(orders, time.Now())

	p, _ := plan.FindPlan(enum.TierPremium, enum.CycleYear)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).
		SetEnvironment(env.Live()).
		SetMembership(member).
		SetWallet(wallet)

	if err := tx.Commit(); err != nil {
		return subscription.PaymentIntent{}, err
	}

	return builder.PaymentIntent()
}
