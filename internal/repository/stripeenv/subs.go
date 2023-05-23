package stripeenv

import (
	"database/sql"
	"errors"

	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// LoadOrFetchSubs tries to retrieve a Stripe subscription from db, and fallback
// to Stripe API if not found.
// Pass `refresh: true` to circumvent db.
func (env Env) LoadOrFetchSubs(id string, refresh bool) (stripe.Subs, error) {
	if !refresh {
		subs, err := env.RetrieveSubs(id)
		if err == nil {
			return subs, nil
		}
	}

	rawSubs, err := env.Client.FetchSubs(id, false)
	if err != nil {
		return stripe.Subs{}, err
	}

	return stripe.NewSubs("", rawSubs), nil
}

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
func (env Env) CreateSubscription(cart reader.ShoppingCart, params stripe.SubsParams) (reader.ShoppingCart, stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return cart, stripe.SubsResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	sugar.Infof("Retrieving membership before creating stripe subscription: %s", cart.Account.CompoundID())

	mmb, err := tx.RetrieveMember(cart.Account.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	cart, err = cart.WithMember(mmb)
	if err != nil {
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	if !cart.Intent.Kind.IsNewSubs() {
		sugar.Errorf("expected shopping cart intent to be new subs, got %s", cart.Intent.Kind)

		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, errors.New("this endpoint only permit creating a new stripe subscription")
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.Client.NewSubs(params.NewSubParams(cart.Account.StripeID.String, cart.StripeItem))

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	sugar.Infof("New subscription created %v", ss.ID)

	subs := stripe.NewSubs(cart.Account.FtcID, ss)

	// Build Membership based on stripe subscription.
	result := stripe.NewSubsBuilder(
		cart,
		subs,
		reader.NewArchiver().ByStripe().WithIntent(cart.Intent.Kind),
	).Build()

	sugar.Infof("A new stripe membership: %+v", result.Member)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		sugar.Info("Inserting stripe membership")
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return cart, stripe.SubsResult{}, err
		}
	} else {
		sugar.Info("Updating an existing membership to stripe")
		err := tx.UpdateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return cart, stripe.SubsResult{}, err
		}
	}

	if !result.CarryOverInvoice.IsZero() {
		sugar.Infof("Saving add-on for after switching to stripe: %v", result.CarryOverInvoice)
		err := tx.SaveInvoice(result.CarryOverInvoice)
		if err != nil {
			_ = tx.Rollback()
			// Since stripe subscription is already created, we should not rollback in case of error.
			sugar.Error(err)
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return cart, stripe.SubsResult{}, err
	}

	return cart, result, nil
}

// UpdateSubscription switches subscription plan.
// It could either change to a different billing cycle, for example from month to year,
// to change to a different product, for example from standard to premium.
// Use cases:
// - Change default payment method;
// - Switch product between standard and premium; between monthly and yearly;
// - Apply coupon. When applying coupon to an existing subscription, the amount will be deducted from
// the next invoice. It won't affect an invoice already generated.
func (env Env) UpdateSubscription(
	subsID string,
	cart reader.ShoppingCart,
	params stripe.SubsParams,
) (reader.ShoppingCart, stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return cart, stripe.SubsResult{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(cart.Account.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, nil
	}

	if !mmb.IsStripeSubsMatch(subsID) {
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, errors.New("mismatched stripe subscription id")
	}

	cart, err = cart.WithMember(mmb)
	if err != nil {
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	if !cart.Intent.Kind.IsUpdating() {
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, errors.New("this endpoint only supports updating an existing valid Stripe subscription")
	}

	currentSubs, err := env.LoadOrFetchSubs(mmb.StripeSubsID.String, false)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	ss, err := env.Client.UpdateSubs(
		currentSubs.ID,
		params.UpdateSubParams(currentSubs.Items[0].ID, cart.StripeItem),
	)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, latest invoice %v", ss.ID, ss.Status, ss.LatestInvoice)

	subs := stripe.NewSubs(cart.Account.FtcID, ss)
	result := stripe.NewSubsBuilder(
		cart,
		subs,
		reader.NewArchiver().ByStripe().WithIntent(cart.Intent.Kind)).
		Build()

	sugar.Infof("Upgraded membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return cart, stripe.SubsResult{}, err
	}

	return cart, result, nil
}

// CancelSubscription cancels a subscription at period end if `CancelParams.Cancel` is true, else reactivate it.
// Here the cancel actually does not delete the subscription.
// It only indicates this subscription won't be automatically renews at period end.
// A canceled subscription is still in active state.
// When stripe says the status is canceled, it means the subscription is expired, deleted, and it won't charge upon period ends.
func (env Env) CancelSubscription(params stripe.CancelParams) (stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	mmb, err := tx.RetrieveMember(params.FtcID)
	sugar.Infof("Current membership cancel/reactivate stripe subscription %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if !mmb.IsStripe() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, sql.ErrNoRows
	}

	if mmb.StripeSubsID.String != params.SubID {
		_ = tx.Rollback()
		return stripe.SubsResult{}, sql.ErrNoRows
	}

	// If you want to cancel it, and membership is not auto-renewal,
	// it means it is already canceled.
	// If cancel is false, you are reactivating a canceled subscription.
	// If the membership is not auto-renewal, it means the member
	// is already reactivated, or not canceled at all.
	// Only cancel and auto-renewal are consistent should you proceed.
	if params.Cancel != mmb.AutoRenewal {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified: false,
			Subs:     stripe.Subs{},
			Member:   mmb,
		}, nil
	}

	ss, err := env.Client.CancelSubs(params.SubID, params.Cancel)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Canceled/reactivated subscription %s, status %s", ss.ID, ss.Status)

	var archiver reader.Archiver
	if params.Cancel {
		archiver = reader.NewArchiver().ByStripe().ActionCancel()
	} else {
		archiver = reader.NewArchiver().ByStripe().ActionReactivate()
	}

	subs := stripe.NewSubs(mmb.FtcID.String, ss)
	result := stripe.SubsSuccessBuilder{
		UserIDs:       mmb.UserIDs,
		Kind:          reader.IntentNull,
		CurrentMember: mmb,
		Subs:          subs,
		Archiver:      archiver,
	}.Build()

	sugar.Infof("Cancelled/reactivated membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Membership canceled/reactivated")

	return result, nil
}
