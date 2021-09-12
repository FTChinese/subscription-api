package subs

import (
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type Invoices struct {
	Purchased   invoice.Invoice
	CarriedOver invoice.Invoice // If carry-over invoice exists, you should add timestamp to old invoices that are not yet completely consumed.
}

// Create new membership based on purchased invoice and optional carry over invoice.
func (i Invoices) membership(userID ids.UserIDs, current reader.Membership) (reader.Membership, error) {
	// This will update expiration date.
	current, err := current.WithInvoice(userID, i.Purchased)

	if err != nil {
		return reader.Membership{}, err
	}

	if i.CarriedOver.IsZero() {
		return current, nil
	}

	// This will update add-on part if carried over invoice is not empty.
	return current.WithInvoice(userID, i.CarriedOver)
}
