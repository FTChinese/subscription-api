package cmsrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// CreateAddOn saves an addon invoice and increase
// membership's addon fields. This is usually done manually
// as a compensation mechanism.
// Do not use it if addon is created automatically.
func (env Env) CreateAddOn(inv invoice.Invoice) (reader.AddOnInvoiceCreated, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.beginAddOnTx()
	if err != nil {
		sugar.Error(err)
		return reader.AddOnInvoiceCreated{}, err
	}
	// Retrieve current membership. It must exist.
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
		Versioned: reader.NewMembershipVersioned(newM).
			WithPriorVersion(member).
			ArchivedBy(reader.NewArchiver().ByManual().ActionAddOn()),
	}, nil
}
