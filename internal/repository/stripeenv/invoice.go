package stripeenv

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (env Env) LoadOrFetchInvoice(id string, refresh bool) (stripe.Invoice, error) {
	if !refresh {
		inv, err := env.RetrieveInvoice(id)
		if err == nil {
			return inv, nil
		}
	}

	rawInv, err := env.Client.FetchInvoice(id)
	if err != nil {
		return stripe.Invoice{}, nil
	}

	return stripe.NewInvoice(rawInv), nil
}
