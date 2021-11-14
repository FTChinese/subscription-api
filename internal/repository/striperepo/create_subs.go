package striperepo

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
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
func (env Env) CreateSubscription(ba account.BaseAccount, item stripe.CheckoutItem, params stripe.SubSharedParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(ba.CompoundID())
	sugar.Infof("Current membership before creating stripe subscription: %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, err := mmb.SubsKindOfStripe(item.Edition())
	//intent, err := reader.NewCheckoutIntents(mmb, params.Edition.Edition).
	//	Get(enum.PayMethodStripe)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot create a new Stripe subscription: " + err.Error(),
			Invalid:    nil,
		}
	}

	sugar.Infof("Stripe subscripiton kind: %s", subsKind.Localize())

	if !subsKind.IsNewSubs() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Only creating new stripe subscription allowed",
			Invalid:    nil,
		}
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.client.NewSubs(
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

	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	// Build Membership based on stripe subscription.
	result, err := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       ba.CompoundIDs(),
		CurrentMember: mmb,
		Kind:          subsKind,
		Action:        reader.ActionActionCreate,
	})

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

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
