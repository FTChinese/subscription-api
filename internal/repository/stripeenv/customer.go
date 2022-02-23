package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
)

// CreateCustomer create a customer under ftc account for user with `ftcID`.
// If reader's current account already have stripe customer id, this action
// is aborted and current reader account is returned.
func (env Env) CreateCustomer(ftcID string) (stripe.Customer, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.BeginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	// Account might not be found, though it is rare.
	baseAccount, err := tx.LockBaseAccount(ftcID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	// If stripe customer id already exists, abort.
	if baseAccount.StripeID.Valid {
		_ = tx.Rollback()
		cus, err := env.LoadOrFetchCustomer(
			baseAccount.StripeID.String,
			false)

		if err != nil {
			return stripe.Customer{}, err
		}

		return cus.WithFtcID(baseAccount.FtcID), nil
	}

	// Request stripe api to create customer.
	// Return *stripe.Error if occurred.
	rawCus, err := env.Client.CreateCustomer(baseAccount.Email)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.Customer{}, err
	}

	cus := stripe.NewCustomer(baseAccount.FtcID, rawCus)

	// Save customer id in our db.
	if err := tx.SaveCustomerID(baseAccount.WithStripeID(cus.ID)); err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	return cus, nil
}

// LoadOrFetchCustomer retrieves customer from our db,
// and fallback to Stripe API if not found.
// NOTE the Customer might not contain ftc id if fetched
// from Stripe.
func (env Env) LoadOrFetchCustomer(id string, refresh bool) (stripe.Customer, error) {
	if !refresh {
		cus, err := env.RetrieveCustomer(id)
		if err == nil {
			return cus, nil
		}
	}

	rawCus, err := env.Client.FetchCustomer(id)
	if err != nil {
		return stripe.Customer{}, err
	}

	return stripe.NewCustomer("", rawCus), nil
}

func (env Env) SetCusDefaultPaymentIfMissing(cusID string, pmID string) (stripe.Customer, error) {
	cus, err := env.LoadOrFetchCustomer(cusID, false)
	if err != nil {
		return stripe.Customer{}, err
	}

	if cus.DefaultPaymentMethodID.Valid {
		return cus, nil
	}

	rawCus, err := env.Client.SetCusDefaultPaymentMethod(cusID, pmID)
	if err != nil {
		return stripe.Customer{}, err
	}

	return stripe.NewCustomer(cus.FtcID, rawCus), nil
}
