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

// LoadOrFetchPaymentMethod retrieve a payment method from
// db, and if not found, fallback to stripe API.
func (env Env) LoadOrFetchPaymentMethod(id string) (stripe.PaymentMethod, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	pi, err := env.RetrievePaymentMethod(id)
	if err == nil {
		return pi, nil
	}
	sugar.Error(err)

	rawPM, err := env.Client.FetchPaymentMethod(id)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	return stripe.NewPaymentMethod(rawPM), nil
}
