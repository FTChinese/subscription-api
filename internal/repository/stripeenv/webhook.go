package stripeenv

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) OnWebhookSubs(subs stripe.Subs, userIDs ids.UserIDs) (stripe.WebhookSubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.WebhookSubsResult{}, err
	}

	// Retrieve membership by stripe subscription id.
	// If found, the membership must be derived from stripe.
	// If not found, two possibilities:
	// 1. No membership exists for this stripe id and ftc id
	// 2. No membership exists for this stripe id. The ftc id
	// side, however, has a membership.
	stripeMmb, err := tx.RetrieveStripeMember(subs.ID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.WebhookSubsResult{}, err
	}

	// Since stripe side does not have membership, concurrency lock won't happen.
	var ftcMmb reader.Membership
	if stripeMmb.IsZero() {
		ftcMmb, err = tx.RetrieveMember(userIDs.CompoundID)
		if err != nil {
			_ = tx.Rollback()
			return stripe.WebhookSubsResult{}, err
		}
	}

	result, err := stripe.WebhookResultBuilder{
		Subs:         subs,
		UserIDs:      userIDs,
		StripeMember: stripeMmb,
		FtcMember:    ftcMmb,
	}.Build()

	if err != nil {
		_ = tx.Rollback()
		return stripe.WebhookSubsResult{}, &stripe.WebhookError{
			ID:        subs.ID,
			EventType: "",
			Message:   err.Error(),
			CurrentStripeMembership: reader.MembershipColumn{
				Membership: stripeMmb,
			},
			CurrentDestMembership: reader.MembershipColumn{
				Membership: ftcMmb,
			},
			TargetUserID: userIDs.CompoundID,
			CreatedUTC:   chrono.TimeNow(),
		}
	}

	// If previously no membership existed
	if result.Versioned.AnteChange.IsZero() {
		err = tx.CreateMember(result.Member)
	} else {
		err = tx.UpdateMember(result.Member)
	}

	if err != nil {
		_ = tx.Rollback()
		return stripe.WebhookSubsResult{}, err
	}

	if !result.CarryOverInvoice.IsZero() {
		err := tx.SaveInvoice(result.CarryOverInvoice)
		if err != nil {
			_ = tx.Rollback()
			return stripe.WebhookSubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return stripe.WebhookSubsResult{}, err
	}

	return result, nil
}
