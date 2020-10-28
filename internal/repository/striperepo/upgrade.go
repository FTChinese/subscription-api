package striperepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
)

// UpgradeSubscription switches subscription plan.
func (env Env) UpgradeSubscription(input stripePkg.SubsInput) (stripePkg.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
	}

	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(input.FtcID),
		UnionID:    null.String{},
	}.MustNormalize())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, nil
	}

	if mmb.IsZero() {
		sugar.Error("membership for stripe upgrading not found")
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, sql.ErrNoRows
	}

	subsKind, err := mmb.StripeSubsKind(input.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}
	// Check whether upgrading is permitted.
	if subsKind != enum.OrderKindUpgrade {
		sugar.Error("upgrading via stripe is not permitted")
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, stripePkg.ErrInvalidStripeSub
	}

	ss, err := input.UpgradeSubs(mmb.StripeSubsID.String)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	newMmb := stripePkg.RefreshMembership(mmb, ss)

	if err := tx.UpdateMember(newMmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
	}

	return stripePkg.SubsResult{
		StripeSubs: ss,
		Member:     newMmb,
		Snapshot:   mmb.Snapshot(reader.ArchiverStripeUpgrade),
	}, nil
}
