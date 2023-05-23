package stripeenv

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SyncSubs syncs stripe subscription in db against Stripe API.
// Used for refresh and webhook notification.
func (env Env) SyncSubs(userIDs ids.UserIDs, subs stripe.Subs, archiver reader.Archiver) (stripe.SubsResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	sugar.Infof("Start syncing stripe subscritpion %s", subs.ID)

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve membership by stripe subscription id.
	// If found, the membership must be derived from stripe.
	// If not found, two possibilities:
	// 1. No membership exists for this stripe id and ftc id
	// 2. No membership exists for this stripe id. The ftc id
	// side, however, has a membership.
	stripeMmb, err := tx.RetrieveStripeMember(subs.ID)
	sugar.Infof("Retrieved membership on stripe side %s", stripeMmb.CompoundID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Since stripe side does not have membership, concurrency lock won't happen.
	var ftcMmb reader.Membership
	if stripeMmb.IsZero() {
		// Try to retrieve membership from FTC side.
		ftcMmb, err = tx.RetrieveMember(userIDs.CompoundID)
		sugar.Infof("Retrieved membership on ftc side %s", ftcMmb.CompoundID)
		if err != nil {
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	// Neither sides of membership might exist.
	result, err := stripe.SubsSyncBuilder{
		UserIDs:      userIDs,
		Subs:         subs,
		StripeMember: stripeMmb,
		FtcMember:    ftcMmb,
		Archiver:     archiver,
	}.Build()
	sugar.Infof("subs modified %t", result.Modified)
	sugar.Infof("membership: %v", result.Member)
	sugar.Infof("versioned: %v", result.Versioned)
	sugar.Infof("carry over: %v", result.CarryOverInvoice)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, &stripe.WebhookError{
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

	// If data is not touched, stop.
	if !result.Modified {
		_ = tx.Rollback()
		sugar.Info("sync result not modified.")
		return result, nil
	}

	// If previously no membership existed
	if result.Versioned.AnteChange.IsZero() {
		err = tx.CreateMember(result.Member)
	} else {
		err = tx.UpdateMember(result.Member)
	}

	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	if !result.CarryOverInvoice.IsZero() {
		sugar.Infof("creating carry over %s, user %s, subscription %s", result.CarryOverInvoice.ID, result.Member.CompoundID, result.Subs.ID)

		err := tx.SaveInvoice(result.CarryOverInvoice)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return stripe.SubsResult{}, err
	}

	return result, nil
}
