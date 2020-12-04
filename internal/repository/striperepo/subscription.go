package striperepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

// CreateSubscription creates a new subscription.
// If performs multiple steps:
// 1. Find existing membership if exist;
// 2. Check what kind of subscription it is, and whether we allow it to continue;
// 3. Create subscription at Stripe API.
// 4. Build FTC membership from Stripe subscription.
// It returns stripe's subscription as is.
//
// error could be:
// util.ErrNonStripeValidSub
// util.ErrStripeDuplicateSub
// util.ErrUnknownSubState
func (env Env) CreateSubscription(co stripe.Checkout) (stripe.CheckoutResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return stripe.CheckoutResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(co.Account.MemberID())

	sugar.Infof("Current membership before creating stripe subscription: %+v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, err := mmb.StripeSubsKind(co.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, err
	}
	if subsKind != enum.OrderKindCreate {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, reader.ErrStripeNotCreate
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.client.CreateSubs(co.StripeParams())

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, err
	}

	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	subs := co.NewSubs(ss)

	// Create Membership based on stripe subscription.
	// Keep existing membership's union id if exists.
	newMmb := co.Membership(subs)
	sugar.Infof("A new stripe membership: %+v", newMmb)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		err := tx.CreateMember(newMmb)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.CheckoutResult{}, err
		}
	} else {
		err := tx.UpdateMember(newMmb)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.CheckoutResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.CheckoutResult{}, err
	}

	payResult, err := stripe.NewPaymentResult(ss)
	if err != nil {
		sugar.Error(err)
	}

	return stripe.CheckoutResult{
		PaymentResult: payResult,
		Subs:          subs,
		Payment:       payResult,
		Member:        newMmb,
		Snapshot:      mmb.Snapshot(reader.ArchiverStripeCreate),
	}, err
}

// SaveSubsError saves any error in stripe response.
func (env Env) SaveSubsError(e stripe.APIError) error {
	_, err := env.db.NamedExec(stripe.StmtSaveAPIError, e)

	if err != nil {
		return err
	}

	return nil
}

// RefreshSubs refresh stripe subscription data if stale.
func (env Env) RefreshSubs(s stripe.Subs) (stripe.CheckoutResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	ss, err := env.client.RetrieveSubs(s.ID)
	if err != nil {
		return stripe.CheckoutResult{}, err
	}

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return stripe.CheckoutResult{}, err
	}

	mmb, err := tx.RetrieveStripeMember(s.ID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, err
	}

	sugar.Infof("Retrieve a stripe member: %+v", mmb)

	newMmb := stripe.RefreshMembership(mmb, ss)

	sugar.Infof("Refreshed membership: %+v", newMmb)

	// update member
	if err := tx.UpdateMember(newMmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CheckoutResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.CheckoutResult{}, err
	}

	sugar.Info("Refresh stripe subscription finished.")
	return stripe.CheckoutResult{
		PaymentResult: stripe.PaymentResult{},
		Subs:          s,
		Payment:       stripe.PaymentResult{},
		Member:        reader.Membership{},
		Snapshot:      reader.MemberSnapshot{},
	}, nil
}

func (env Env) UpsertSubs(s stripe.Subs) error {
	_, err := env.db.NamedExec(stripe.StmtInsertSubs, s)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveSubs(id string) (stripe.Subs, error) {
	var s stripe.Subs
	err := env.db.Get(&s, stripe.StmtRetrieveSubs, id)
	if err != nil {
		return s, err
	}

	return s, nil
}
