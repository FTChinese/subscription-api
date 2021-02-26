package striperepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/cart"
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
func (env Env) CreateSubscription(params stripe.SubsParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(params.Account.MemberID())
	sugar.Infof("Current membership before creating stripe subscription: %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	intent, err := cart.NewCheckoutIntents(mmb, params.Edition.Edition).
		Get(enum.PayMethodStripe)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot create a new Stripe subscription: " + err.Error(),
			Invalid:    nil,
		}
	}

	if !intent.IsNewStripe() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Only creating new stripe subscription allowed",
			Invalid:    nil,
		}
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.client.CreateSubs(params)

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
		UserIDs:       params.Account.MemberID(),
		CurrentMember: mmb,
		Kind:          intent.SubsKind,
		Action:        reader.ActionCreate,
	})

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("A new stripe membership: %+v", result.Member)

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	} else {
		err := tx.UpdateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if !result.AddOn.IsZero() {
		if err := tx.SaveAddOn(result.AddOn); err != nil {
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
