package striperepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

// UpdateSubscription switches subscription plan.
func (env Env) UpdateSubscription(cfg stripe.SubsParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(cfg.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, nil
	}

	intent, err := reader.NewCheckoutIntents(mmb, cfg.Edition.Edition).
		Get(enum.PayMethodStripe)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot perform updating Stripe subscription: " + err.Error(),
			Invalid:    nil,
		}
	}

	if !intent.IsUpdatingStripe() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, &render.ResponseError{
			StatusCode: http.StatusBadGateway,
			Message:    "This endpoint only support updating an existing valid Stripe subscription while you can only " + intent.Description(),
			Invalid:    nil,
		}
	}

	ss, err := env.client.UpdateSubs(mmb.StripeSubsID.String, cfg)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	result, err := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       mmb.MemberID,
		Kind:          intent.SubsKind,
		CurrentMember: mmb,
		Action:        reader.ActionUpgrade,
	})

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

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
