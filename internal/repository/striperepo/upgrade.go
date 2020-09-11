package striperepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go"
)

// UpgradeStripeSubs switches subscription plan.
func (env Env) UpgradeSubscription(input stripePkg.SubsInput) (*stripeSdk.Subscription, error) {
	defer logger.Sync()
	sugar := logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(input.FtcID),
		UnionID:    null.String{},
	}.MustNormalize())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, nil
	}

	if mmb.IsZero() {
		sugar.Error("membership for stripe upgrading not found")
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	subsKind, err := mmb.StripeSubsKind(input.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}
	// Check whether upgrading is permitted.
	if subsKind != enum.OrderKindUpgrade {
		sugar.Error("upgrading via stripe is not permitted")
		_ = tx.Rollback()
		return nil, stripePkg.ErrInvalidStripeSub
	}

	ss, err := input.UpgradeSubs(mmb.StripeSubsID.String)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	mmb = stripePkg.RefreshMembership(mmb, ss)

	if err := tx.UpdateMember(mmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return nil, err
	}

	return ss, nil
}
