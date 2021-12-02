package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

// CreateCustomer create a customer under ftc account for user with `ftcID`.
// If reader's current account already have stripe customer id, this action
// is aborted and current reader account is returned.
func (env Env) CreateCustomer(ftcID string) (stripe.CustomerAccount, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.CustomerAccount{}, err
	}

	// Account might not be found, though it is rare.
	baseAccount, err := tx.BaseAccountForStripe(ftcID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.CustomerAccount{}, err
	}

	// If stripe customer id already exists, abort.
	if baseAccount.StripeID.Valid {
		_ = tx.Rollback()
		cus, err := env.Client.RetrieveCustomer(baseAccount.StripeID.String)
		if err != nil {
			return stripe.CustomerAccount{}, err
		}

		return stripe.NewCustomerAccount(baseAccount, cus), nil
	}

	// Request stripe api to create customer.
	// Return *stripe.Error if occurred.
	cus, err := env.Client.CreateCustomer(baseAccount.Email)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.CustomerAccount{}, err
	}

	ca := stripe.NewCustomerAccount(baseAccount, cus)

	// Save customer id in our db.
	// There might be SQL errors.
	if err := tx.SaveCustomerID(ca.BaseAccount); err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return stripe.CustomerAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.CustomerAccount{}, err
	}

	return ca, nil
}

// SetCustomer set stripe customer id.
// Used when creating checkout session.
func (env Env) SetCustomer(a account.BaseAccount) error {
	_, err := env.DBs.Write.NamedExec(
		stripe.StmtSetCustomerID,
		a)

	if err != nil {
		return err
	}

	return nil
}
