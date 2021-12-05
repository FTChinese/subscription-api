package striperepo

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

// UpdateSubscription switches subscription plan.
// It could either change to a different billing cycle, for example from month to year,
// to change to a different product, for example from standard to premium.
func (env Env) UpdateSubscription(ba account.BaseAccount, item stripe.CheckoutItem, params stripe.SubSharedParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginStripeTx()
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

	ss, err := env.client.GetSubs(mmb.StripeSubsID.String, false)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	ss, err = env.client.UpdateSubs(
		ss.ID,
		item.UpdateSubParams(ss, params),
	)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	result, err := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       mmb.UserIDs,
		Kind:          subsKind,
		CurrentMember: mmb,
		Action:        reader.ArchiveActionUpgrade,
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
