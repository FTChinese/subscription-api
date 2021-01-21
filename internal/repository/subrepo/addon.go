package subrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

func (env Env) TransferAddOn(ids reader.MemberID) (subs.AddOnConsumed, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.AddOnConsumed{}, err
	}

	sugar.Info("Start retrieving membership for %v", ids)

	member, err := otx.RetrieveMember(ids)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	if !member.ShouldUseAddOn() {
		sugar.Info("Add on cannot be transferred to membership %v", member)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	addOns, err := otx.ListAddOn(member.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	if len(addOns) == 0 {
		sugar.Info("No add-on")
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, sql.ErrNoRows
	}

	// otherwise we might override valid data.
	result := subs.ConsumeAddOns(addOns, member)

	err = otx.AddOnsConsumed(result.AddOnIDs.ToSlice())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	err = otx.UpdateMember(result.Membership)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	return result, nil
}
