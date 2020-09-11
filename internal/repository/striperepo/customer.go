package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
)

// CreateStripeCustomer create a customer under ftc account for user with `ftcID`.
// If readers current account already have stripe customer id, this action
// is aborted and current reader account is returned.
// It is done in a transaction so that we won't create duplicate customer
// for the same reader.
func (env Env) CreateStripeCustomer(ftcID string) (reader.FtcAccount, error) {
	defer logger.Sync()
	sugar := logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		sugar.Error(err)
		return reader.FtcAccount{}, err
	}

	// Account might not be found, though it is rare.
	account, err := tx.RetrieveAccount(ftcID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return reader.FtcAccount{}, err
	}

	// If stripe customer id already exists, abort.
	if account.StripeID.Valid {
		sugar.Error(err)
		_ = tx.Rollback()
		return account, nil
	}

	// Request stripe api to create customer.
	// Return *stripe.Error if occurred.
	stripeID, err := stripe.CreateCustomer(account.Email)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return reader.FtcAccount{}, err
	}

	// Add stripe customer id to current account.
	account.StripeID = null.StringFrom(stripeID)

	// Save customer id in our db.
	// There might be SQL errors.
	if err := tx.SavedStripeID(account); err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return reader.FtcAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return reader.FtcAccount{}, err
	}

	return account, nil
}
