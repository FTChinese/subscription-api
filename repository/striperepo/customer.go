package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

// createCustomer sends request to customer creation endpoint.
func createCustomer(email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)

	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

// CreateStripeCustomer create a customer under ftc account
// for user with `ftcID`.
func (env StripeEnv) CreateStripeCustomer(ftcID string) (reader.Account, error) {
	log := logger.WithField("trace", "StripeEnv.CreateStripeCustomer")

	tx, err := env.beginAccountTx()
	if err != nil {
		log.Error(err)
		return reader.Account{}, err
	}

	account, err := tx.RetrieveAccount(ftcID)
	if err != nil {
		_ = tx.Rollback()
		log.Error(err)
		return reader.Account{}, err
	}

	if account.StripeID.Valid {
		_ = tx.Rollback()
		return account, nil
	}

	stripeID, err := createCustomer(account.Email)
	if err != nil {
		_ = tx.Rollback()
		log.Error(err)
		return reader.Account{}, err
	}

	account.StripeID = null.StringFrom(stripeID)

	if err := tx.SavedStripeID(account); err != nil {
		_ = tx.Rollback()
		log.Error(err)
		return reader.Account{}, err
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return reader.Account{}, err
	}

	return account, nil
}
