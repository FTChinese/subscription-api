package subrepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) RedeemAddOn(ids reader.MemberID) (reader.AddOnConsumed, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return reader.AddOnConsumed{}, err
	}

	sugar.Infof("Start retrieving membership for %v", ids)

	member, err := otx.RetrieveMember(ids)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, err
	}

	sugar.Infof("Membership retrieved %v", member)

	if member.IsZero() {
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, errors.New("")
	}

	if err := member.ShouldUseAddOn(); err != nil {
		sugar.Info("Add on cannot be transferred to membership %v", member)
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, err
	}

	addOns, err := otx.ListAddOn(member.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, err
	}

	if len(addOns) == 0 {
		sugar.Info("No add-on")
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, sql.ErrNoRows
	}

	// otherwise we might override valid data.
	result := member.ConsumeAddOns(addOns)

	err = otx.AddOnsConsumed(result.AddOnIDs.ToSlice())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, err
	}

	err = otx.UpdateMember(result.Membership)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnConsumed{}, err
	}

	return result, nil
}
