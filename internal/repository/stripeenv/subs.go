package stripeenv

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/account"
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
func (env Env) CreateSubscription(
	ba account.BaseAccount,
	item stripe.CheckoutItem,
	params stripe.SubsParams,
) (stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	sugar.Infof("Retrieving membership before creating stripe subscription: %s", ba.CompoundID())
	mmb, err := tx.RetrieveMember(ba.CompoundID())

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, err := mmb.SubsKindOfStripe(item.Price.Edition())
	//intent, err := reader.NewCheckoutIntents(mmb, params.Edition.Edition).
	//	Get(enum.PayMethodStripe)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Stripe subscripiton kind: %s", subsKind.Localize())

	if !subsKind.IsNewSubs() {
		sugar.Infof("Abort stripe new subscription due to not eligible.")
		_ = tx.Rollback()
		return stripe.SubsResult{}, errors.New("this endpoint only permit creating a new stripe subscription")
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.Client.NewSubs(
		item.NewSubParams(ba.StripeID.String, params),
	)

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("New subscription created %v", ss.ID)

	// Build Membership based on stripe subscription.
	result := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       ba.CompoundIDs(),
		CurrentMember: mmb,
		Kind:          subsKind,
		Action:        reader.ArchiveActionCreate,
	})

	sugar.Infof("A new stripe membership: %+v", result.Member)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		sugar.Info("Inserting stripe membership")
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	} else {
		sugar.Info("Updating an existing membership to stripe")
		err := tx.UpdateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if !result.CarryOverInvoice.IsZero() {
		sugar.Infof("Saving add-on for after switching to stripe: %v", result.CarryOverInvoice)
		if err := tx.SaveInvoice(result.CarryOverInvoice); err != nil {
			// Since stripe subscription is already created, we should not rollback in case of error.
			sugar.Error(err)
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
}

// UpdateSubscription switches subscription plan.
// It could either change to a different billing cycle, for example from month to year,
// to change to a different product, for example from standard to premium.
func (env Env) UpdateSubscription(
	ba account.BaseAccount,
	item stripe.CheckoutItem,
	params stripe.SubsParams,
) (stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(ba.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, nil
	}

	subsKind, err := mmb.SubsKindOfStripe(item.Price.Edition())

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot perform updating Stripe subscription: " + err.Error(),
			Invalid:    nil,
		}
	}

	if !subsKind.IsUpdating() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadGateway,
			Message:    "This endpoint only support updating an existing valid Stripe subscription",
			Invalid:    nil,
		}
	}

	ss, err := env.Client.FetchSubs(mmb.StripeSubsID.String, false)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	ss, err = env.Client.UpdateSubs(
		ss.ID,
		item.UpdateSubParams(ss, params),
	)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
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
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
}

// RefreshSubscription save stripe subscription and optionally update membership linked to it.
func (env Env) RefreshSubscription(ss *sdk.Subscription, param stripe.SubsResultParams) (stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar().
		With("webhook", "stripe-subscription").
		With("id", ss.ID)

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership by ftc id.
	// If current membership is empty, we should create it.
	currMmb, err := tx.RetrieveMember(param.UserIDs.CompoundID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	param.CurrentMember = currMmb

	result := stripe.NewSubsResult(ss, param)

	// Ensure that current membership is created via stripe.
	if !result.Subs.ShouldUpsert(currMmb) {
		_ = tx.Rollback()
		sugar.Infof("Stripe subscription cannot update/insert its membership")
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 result.Subs,
			Member:               currMmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// If nothing changed.
	if !result.Member.IsModified(currMmb) {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 result.Subs,
			Member:               result.Member,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// Insert to update membership.
	if currMmb.IsZero() {
		if err := tx.CreateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	} else {
		if err := tx.UpdateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
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

	// If you want to cancel it, and membership is not auto renewal,
	// it means it is already canceled.
	// If cancel is false, you are reactivating a canceled subscription.
	// If the membership is not auto renewal, it means the member
	// is already reactivated, or not canceled at all.
	// Only cancel and auto renewal are consistent should you proceed.
	if params.Cancel != mmb.AutoRenewal {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 stripe.Subs{},
			Member:               mmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	ss, err := env.Client.CancelSubs(params.SubID, params.Cancel)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
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
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Membership canceled/reactivated")

	return result, nil
}
