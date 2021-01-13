package subrepo

import "github.com/FTChinese/subscription-api/pkg/subs"

func (env Env) TransferAddOn(ids []string) (subs.AddOnConsumed, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.AddOnConsumed{}, err
	}

	sugar.Info("Start retrieving membership for %v", ids)

	member, err := otx.RetrieveMemberV2(ids)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	if !member.ShouldUseAddOn() {
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	addOns, err := otx.ListAddOn(member.MemberID)
	if err != nil {
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	// otherwise we might override valid data.
	result := subs.TransferAddOn(addOns, member)

	err = otx.AddOnsConsumed(subs.GetAddOnIDs(result.AddOns))
	if err != nil {
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	err = otx.UpdateMember(result.Membership)
	if err != nil {
		_ = otx.Rollback()
		return subs.AddOnConsumed{}, err
	}

	return result, nil
}
