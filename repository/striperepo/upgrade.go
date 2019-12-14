package striperepo

import (
	"database/sql"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	ftcStripe "gitlab.com/ftchinese/subscription-api/models/stripe"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

func upgradeSub(p ftcStripe.SubParams, subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(p.GetStripePlanID()),
			},
		},
	}

	params.AddExpand("latest_invoice.payment_intent")

	if p.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(p.IdempotencyKey)
	}

	if p.Coupon.Valid {
		params.Coupon = stripe.String(p.DefaultPaymentMethod.String)
	}

	if p.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(p.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(p.IdempotencyKey)
	return sub.Update(subID, params)
}

// UpgradeStripeSubs switches subscription plan.
func (env StripeEnv) UpgradeSubscription(
	id reader.MemberID,
	params ftcStripe.SubParams,
) (*stripe.Subscription, error) {

	log := logger.WithField("trace", "StripeEnv.UpgradeSubscription")

	tx, err := env.beginOrderTx()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, nil
	}

	if mmb.IsZero() {
		log.Error("membership for stripe upgrading not found")
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	// Check whether upgrading is permitted.
	if !mmb.PermitStripeUpgrade() {
		log.Error("upgrading via stripe is not permitted")
		_ = tx.Rollback()
		return nil, util.ErrInvalidStripeSub
	}

	ss, err := upgradeSub(params, mmb.StripeSubID.String)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	mmb = ftcStripe.RefreshMembership(mmb, ss)

	if err := tx.UpdateMember(mmb); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ss, nil
}
