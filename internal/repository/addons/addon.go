package addons

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ClaimAddOn uses the user's addon invoices to extend expiration date.
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

	// Check if addon should be used;
	// otherwise we might override valid data.
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

	// Perform the transfer from addon to expiration date.
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

func (env Env) CreateAddOn(inv invoice.Invoice) (reader.AddOnInvoiceCreated, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.beginAddOnTx()
	if err != nil {
		sugar.Error(err)
		return reader.AddOnInvoiceCreated{}, err
	}
	// Retrieve current membership. It must exists.
	member, err := otx.RetrieveMember(inv.CompoundID)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnInvoiceCreated{}, err
	}
	if member.IsZero() {
		sugar.Error("CreateAddOn: membership not found")
		_ = otx.Rollback()
		return reader.AddOnInvoiceCreated{}, sql.ErrNoRows
	}

	sugar.Infof("Membership retrieved %v", member)

	newM := member.PlusAddOn(addon.New(inv.Tier, inv.TotalDays()))

	err = otx.SaveInvoice(inv)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnInvoiceCreated{}, err
	}

	err = otx.UpdateMember(newM)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return reader.AddOnInvoiceCreated{}, err
	}

	if err := otx.Commit(); err != nil {
		return reader.AddOnInvoiceCreated{}, err
	}

	return reader.AddOnInvoiceCreated{
		Invoice:    inv,
		Membership: newM,
		Snapshot:   member.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn)),
	}, nil
}
