package stripeenv

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

// FetchAndSavePaymentMethod fetches a payment method from
// stripe and save it to db.
func (env Env) FetchAndSavePaymentMethod(pmID string) (stripe.PaymentMethod, error) {
	rawPM, err := env.Client.FetchPaymentMethod(pmID)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	pm := stripe.NewPaymentMethod(rawPM)

	err = env.UpsertPaymentMethod(pm)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	return pm, nil
}
