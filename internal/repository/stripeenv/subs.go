package stripeenv

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	sdk "github.com/stripe/stripe-go/v72"
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
func (env Env) CreateSubscription(cart reader.ShoppingCart, params stripe.SubsParams) (reader.ShoppingCart, stripe.SubsSuccess, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return cart, stripe.SubsSuccess{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	sugar.Infof("Retrieving membership before creating stripe subscription: %s", cart.Account.CompoundID())

	mmb, err := tx.RetrieveMember(cart.Account.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}

	cart, err = cart.WithMember(mmb)
	if err != nil {
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}

	if !cart.Intent.Kind.IsNewSubs() {
		sugar.Errorf("expected shopping cart intent to be new subs, got %s", cart.Intent.Kind)

		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, errors.New("this endpoint only permit creating a new stripe subscription")
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
		return cart, stripe.SubsSuccess{}, err
	}

	sugar.Infof("New subscription created %v", ss.ID)

	// Build Membership based on stripe subscription.
	result := stripe.NewSubsResultBuilder(
		cart,
		reader.NewArchiver().ByStripe().WithIntent(cart.Intent.Kind),
	).Build(ss)

	sugar.Infof("A new stripe membership: %+v", result.Member)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		sugar.Info("Inserting stripe membership")
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return cart, stripe.SubsSuccess{}, err
		}
	} else {
		sugar.Info("Updating an existing membership to stripe")
		err := tx.UpdateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return cart, stripe.SubsSuccess{}, err
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
		return cart, stripe.SubsSuccess{}, err
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
) (reader.ShoppingCart, stripe.SubsSuccess, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return cart, stripe.SubsSuccess{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(cart.Account.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, nil
	}

	if !mmb.IsStripeSubsMatch(subsID) {
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, errors.New("mismatched stripe subscription id")
	}

	cart, err = cart.WithMember(mmb)
	if err != nil {
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}

	if !cart.Intent.Kind.IsUpdating() {
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, errors.New("this endpoint only supports updating an existing valid Stripe subscription")
	}

	currentSubs, err := env.LoadOrFetchSubs(mmb.StripeSubsID.String, false)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}

	ss, err := env.Client.UpdateSubs(
		currentSubs.ID,
		params.UpdateSubParams(currentSubs.Items[0].ID, cart.StripeItem),
	)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	result := stripe.NewSubsResultBuilder(
		cart,
		reader.NewArchiver().ByStripe().WithIntent(cart.Intent.Kind)).
		Build(ss)

	sugar.Infof("Upgraded membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return cart, stripe.SubsSuccess{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return cart, stripe.SubsSuccess{}, err
	}

	return cart, result, nil
}

// RefreshSubscription save stripe subscription and optionally update membership linked to it.
func (env Env) RefreshSubscription(
	ss *sdk.Subscription,
	ba account.BaseAccount,
) (stripe.SubsSuccess, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar().
		With("webhook", "stripe-subscription").
		With("id", ss.ID)

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	// Retrieve current membership by ftc id.
	// If current membership is empty, we should create it.
	currMmb, err := tx.RetrieveMember(ba.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	result := stripe.SubsResultBuilder{
		UserIDs:       ba.CompoundIDs(),
		Kind:          reader.IntentNull,
		CurrentMember: currMmb,
		Archiver:      reader.NewArchiver().ByStripe().ActionRefresh(),
	}.Build(ss)

	// Ensure that current membership is created via stripe and data actually changed.
	// if current membership turned to alpay/wxpay/apple, we should stop.
	if !result.Modified {
		_ = tx.Rollback()
		return result, nil
	}

	// Insert to update membership.
	if currMmb.IsZero() {
		if err := tx.CreateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsSuccess{}, err
		}
	} else {
		if err := tx.UpdateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsSuccess{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	return result, nil
}

// CancelSubscription cancels a subscription at period end if `CancelParams.Cancel` is true, else reactivate it.
// Here the cancel actually does not delete the subscription.
// It only indicates this subscription won't be automatically renews at period end.
// A canceled subscription is still in active state.
// When stripe says the status is canceled, it means the subscription is expired, deleted, and it won't charge upon period ends.
func (env Env) CancelSubscription(params stripe.CancelParams) (stripe.SubsSuccess, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	mmb, err := tx.RetrieveMember(params.FtcID)
	sugar.Infof("Current membership cancel/reactivate stripe subscription %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	if !mmb.IsStripe() {
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, sql.ErrNoRows
	}

	if mmb.StripeSubsID.String != params.SubID {
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, sql.ErrNoRows
	}

	// If you want to cancel it, and membership is not auto-renewal,
	// it means it is already canceled.
	// If cancel is false, you are reactivating a canceled subscription.
	// If the membership is not auto-renewal, it means the member
	// is already reactivated, or not canceled at all.
	// Only cancel and auto-renewal are consistent should you proceed.
	if params.Cancel != mmb.AutoRenewal {
		_ = tx.Rollback()
		return stripe.SubsSuccess{
			Modified: false,
			Subs:     stripe.Subs{},
			Member:   mmb,
		}, nil
	}

	ss, err := env.Client.CancelSubs(params.SubID, params.Cancel)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	sugar.Infof("Canceled/reactivated subscription %s, status %s", ss.ID, ss.Status)

	var archiver reader.Archiver
	if params.Cancel {
		archiver = reader.NewArchiver().ByStripe().ActionCancel()
	} else {
		archiver = reader.NewArchiver().ByStripe().ActionReactivate()
	}

	result := stripe.SubsResultBuilder{
		UserIDs:       mmb.UserIDs,
		Kind:          reader.IntentNull,
		CurrentMember: mmb,
		Archiver:      archiver,
	}.Build(ss)

	sugar.Infof("Cancelled/reactivated membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	sugar.Infof("Membership canceled/reactivated")

	return result, nil
}
