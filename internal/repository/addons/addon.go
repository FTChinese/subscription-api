package addons

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) ClaimAddOn(ids pkg.UserIDs) (reader.AddOnClaimed, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.beginAddOnTx()
	if err != nil {
		sugar.Error(err)
		return reader.AddOnClaimed{}, err
	}

	sugar.Infof("Start retrieving membership for %v", ids)

	// Retrieve current membership. It must exists.
	member, err := otx.RetrieveMember(ids.CompoundID)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	sugar.Infof("Membership retrieved %v", member)

	if err := member.ShouldUseAddOn(); err != nil {
		sugar.Infof("Add on cannot be transferred to membership %v", member)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	// List all add-on invoices that is not consumed yet.
	addOns, err := otx.AddOnInvoices(member.UserIDs)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnClaimed{}, err
	}

	sugar.Infof("Add-on invoices: %v", addOns)

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

	err = otx.UpdateMember(result.Membership)
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

	if err := otx.Commit(); err != nil {
		return reader.AddOnClaimed{}, err
	}

	return result, nil
}
