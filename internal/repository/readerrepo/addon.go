package readerrepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) ClaimAddOn(ids reader.MemberID) (reader.AddOnClaimed, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return reader.AddOnClaimed{}, err
	}

	sugar.Infof("Start retrieving membership for %v", ids)

	member, err := otx.RetrieveMember(ids)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	sugar.Infof("Membership retrieved %v", member)

	if member.IsZero() {
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, errors.New("")
	}

	if err := member.ShouldUseAddOn(); err != nil {
		sugar.Info("Add on cannot be transferred to membership %v", member)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	addOns, err := otx.AddOnInvoices(member.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	if len(addOns) == 0 {
		sugar.Info("No add-on")
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, sql.ErrNoRows
	}

	// otherwise we might override valid data.
	result, err := member.ClaimAddOns(addOns)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	for _, inv := range result.Invoices {
		err = otx.AddOnInvoiceConsumed(inv)
		if err != nil {
			sugar.Error(err)
			_ = otx.Rollback()
			return reader.AddOnClaimed{}, err
		}
	}

	err = otx.UpdateMember(result.Membership)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	return result, nil
}
