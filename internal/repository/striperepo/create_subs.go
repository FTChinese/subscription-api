package striperepo

import (
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
func (env Env) CreateSubscription(cfg stripe.PaymentConfig) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(cfg.Account.MemberID())
	sugar.Infof("Current membership before creating stripe subscription: %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	err = mmb.PermitStripeCreate(cfg.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := env.client.CreateSubs(cfg)

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
	result, err := stripe.SubsBuilder{
		IDs:           cfg.Account.MemberID(),
		CurrentMember: mmb,
		SS:            ss,
		Action:        reader.ActionCreate,
	}.Build()

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

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
}
