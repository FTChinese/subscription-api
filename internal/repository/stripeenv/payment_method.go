package stripeenv

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

// LoadOrFetchPaymentMethod retrieve a payment method from
// db, and if not found, fallback to stripe API.
func (env Env) LoadOrFetchPaymentMethod(id string, refresh bool) (stripe.PaymentMethod, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	// If not refreshing, use local db.
	if !refresh {
		pi, err := env.RetrievePaymentMethod(id)
		// If found, return it directly.
		if err == nil {
			return pi, nil
		}

		// otherwise, fallthrough.
		sugar.Error(err)
	}

	// If refreshing, or not found in local db.
	rawPM, err := env.Client.FetchPaymentMethod(id)
	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	return stripe.NewPaymentMethod(rawPM), nil
}

func (env Env) LoadOrFetchSetupIntent(id string, refresh bool) (stripe.SetupIntent, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	// If not refreshing, use local db.
	if !refresh {
		si, err := env.RetrieveSetupIntent(id)
		// If found, return it directly.
		if err == nil {
			return si, nil
		}

		// otherwise, fallthrough.
		sugar.Error(err)
	}

	// If refreshing, or not found in local db.
	rawSI, err := env.Client.FetchSetupIntent(id, false)
	if err != nil {
		return stripe.SetupIntent{}, err
	}

	return stripe.NewSetupIntent(rawSI), nil
}
