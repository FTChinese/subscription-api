package stripeenv

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	sdk "github.com/stripe/stripe-go/v72"
	"net/http"
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
func (env Env) CreateSubscription(cart pw.ShoppingCart, params pw.StripeSubsParams) (stripe.SubsSuccess, pw.ShoppingCart, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, cart, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	sugar.Infof("Retrieving membership before creating stripe subscription: %s", cart.Account.CompoundID())

	mmb, err := tx.RetrieveMember(cart.Account.CompoundID())

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, cart, err
	}

	cart = cart.WithMember(mmb)

	if cart.Intent.Kind != reader.SubsKindCreate {
		sugar.Errorf("expected shopping cart intent to be %s, received %s", reader.SubsKindCreate, cart.Intent.Kind)

		_ = tx.Rollback()
		return stripe.SubsSuccess{}, cart, errors.New("this endpoint only permit creating a new stripe subscription")
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.Client.NewSubs(
		cart.StripeItem.
			NewSubParams(cart.Account.StripeID.String, params),
	)

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, cart, err
	}

	sugar.Infof("New subscription created %v", ss.ID)

	// Build Membership based on stripe subscription.
	result := stripe.NewSubsCreated(cart, ss)

	sugar.Infof("A new stripe membership: %+v", result.Member)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		sugar.Info("Inserting stripe membership")
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsSuccess{}, cart, err
		}
	} else {
		sugar.Info("Updating an existing membership to stripe")
		err := tx.UpdateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsSuccess{}, cart, err
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
		return stripe.SubsSuccess{}, cart, err
	}

	return result, cart, nil
}

// UpdateSubscription switches subscription plan.
// It could either change to a different billing cycle, for example from month to year,
// to change to a different product, for example from standard to premium.
func (env Env) UpdateSubscription(
	ba account.BaseAccount,
	item pw.CartItemStripe,
	params pw.StripeSubsParams,
) (stripe.SubsSuccess, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(ba.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, nil
	}

	subsKind, err := mmb.SubsKindOfStripe(item.Recurring.Edition())

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot perform updating Stripe subscription: " + err.Error(),
			Invalid:    nil,
		}
	}

	if !subsKind.IsUpdating() {
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, &render.ResponseError{
			StatusCode: http.StatusBadGateway,
			Message:    "This endpoint only support updating an existing valid Stripe subscription",
			Invalid:    nil,
		}
	}

	ss, err := env.Client.FetchSubs(mmb.StripeSubsID.String, false)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	ss, err = env.Client.UpdateSubs(
		ss.ID,
		item.UpdateSubParams(ss, params),
	)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	result := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       mmb.UserIDs,
		Kind:          subsKind,
		CurrentMember: mmb,
		Action:        reader.ArchiveActionUpgrade,
	})

	sugar.Infof("Upgraded membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsSuccess{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsSuccess{}, err
	}

	return result, nil
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

	result := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       ba.CompoundIDs(),
		Kind:          reader.SubsKindRefreshAutoRenew,
		CurrentMember: currMmb,
		Action:        reader.ArchiveActionRefresh,
	})

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

	var action reader.ArchiveAction
	if params.Cancel {
		action = reader.ArchiveActionCancel
	} else {
		action = reader.ArchiveActionReactivate
	}

	result := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       mmb.UserIDs,
		CurrentMember: mmb,
		Action:        action,
	})

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
