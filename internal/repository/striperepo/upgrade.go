package striperepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
)

// UpgradeSubscription switches subscription plan.
func (env Env) UpgradeSubscription(co ftcStripe.Checkout) (ftcStripe.CheckoutResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return ftcStripe.CheckoutResult{}, err
	}

	mmb, err := tx.RetrieveMember(co.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, nil
	}
	if mmb.IsZero() {
		sugar.Error("membership for stripe upgrading not found")
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, sql.ErrNoRows
	}
	if mmb.StripeSubsID.IsZero() {
		sugar.Error("Cannot upgrade a non-stripe membership with stripe")
		return ftcStripe.CheckoutResult{}, errors.New("not a stripe member")
	}

	subsKind, err := mmb.StripeSubsKind(co.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, err
	}
	// Check whether upgrading is permitted.
	if subsKind != enum.OrderKindUpgrade {
		sugar.Error("Not upgrading request")
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, ftcStripe.ErrInvalidStripeSub
	}

	ss, err := env.client.UpgradeSubs(mmb.StripeSubsID.String, co.StripeParams())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	subs := co.NewSubs(ss)

	newMmb := co.Membership(subs)

	if err := tx.UpdateMember(newMmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return ftcStripe.CheckoutResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return ftcStripe.CheckoutResult{}, err
	}

	payResult, err := ftcStripe.NewPaymentResult(ss)
	if err != nil {
		return ftcStripe.CheckoutResult{}, err
	}

	return ftcStripe.CheckoutResult{
		PaymentResult: payResult,
		StripeSubs:    ss,
		Subs:          subs,
		Payment:       payResult,
		Member:        newMmb,
		Snapshot:      mmb.Snapshot(reader.ArchiverStripeUpgrade),
	}, nil
}
