package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/jmoiron/sqlx"
)

type AddOnTx struct {
	SharedTx
}

func NewAddOnTx(tx *sqlx.Tx) AddOnTx {
	return AddOnTx{
		SharedTx: NewSharedTx(tx),
	}
}

// AddOnExistsForOrder checks if the specified order already created an invoice.
func (tx AddOnTx) AddOnExistsForOrder(orderID string) (bool, error) {
	var ok bool
	err := tx.Get(&ok, invoice.StmtAddOnExistsForOrder, orderID)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// AddOnInvoices retrieves all invoices of type add_on that are not consumed yet.
// This is used to migrate addons to current membership.
func (tx AddOnTx) AddOnInvoices(ids ids.UserIDs) ([]invoice.Invoice, error) {
	var inv []invoice.Invoice
	err := tx.Select(&inv, invoice.StmtListAddOnInvoiceLock, ids.BuildFindInSet())
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (tx AddOnTx) AddOnInvoiceConsumed(inv invoice.Invoice) error {
	_, err := tx.NamedExec(invoice.StmtAddOnInvoiceConsumed, inv)
	if err != nil {
		return err
	}

	return nil
}
