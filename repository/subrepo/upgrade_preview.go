package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"time"
)

// PreviewUpgrade calculates how much should a user to pay
// to perform upgrading.
func (env Env) PreviewUpgrade(userID reader.MemberID, plan product.ExpandedPlan) (subs.PaymentIntent, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subs.PaymentIntent{}, err
	}

	// Retrieve existing membership to see if user is valid
	// to upgrade.
	member, err := tx.RetrieveMember(userID)
	if err != nil {
		_ = tx.Rollback()
		return subs.PaymentIntent{}, err
	}

	orderKind, err := member.AliWxSubsKind(product.NewStdYearEdition())
	if err != nil {
		_ = tx.Rollback()
		return subs.PaymentIntent{}, err
	}

	// If user is not qualified to upgrade, deny it.
	if orderKind != enum.OrderKindUpgrade {
		_ = tx.Rollback()
		return subs.PaymentIntent{}, subs.ErrUpgradeInvalid
	}

	// Retrieve all orders with balance remaining
	orders, err := tx.FindBalanceSources(userID)
	if err != nil {
		_ = tx.Rollback()
		return subs.PaymentIntent{}, err
	}

	// Calculates the balance of user's wallet.
	wallet := subs.NewWallet(orders, time.Now())

	orderBuilder := subs.NewOrderBuilder(userID).
		SetPlan(plan).
		SetEnvConfig(env.BuildConfig).
		SetWallet(wallet)

	if err := tx.Commit(); err != nil {
		return subs.PaymentIntent{}, err
	}

	return orderBuilder.PaymentIntent()
}
