package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/pkg/apple"
)

func (env IAPEnv) VerifyReceipt(r apple.VerificationRequestBody) (apple.VerificationResponseBody, error) {
	pw, err := getReceiptPassword()
	if err != nil {
		return apple.VerificationResponseBody{}, err
	}

	r.Password = pw
	r.ExcludeOldTransactions = false

	url := env.c.GetReceiptVerificationURL()
	logger.
		WithField("trace", "IAPEnv.VerifyReceipt").
		WithField("Using verification url %s", url)

	_, body, errs := request.
		Post(url).
		Send(r).End()

	if errs != nil {
		logger.WithField("trace", "IAPEnv.VerifyReceipt").Error(errs)
		return apple.VerificationResponseBody{}, errs[0]
	}

	logger.WithField("trace", "IAPEnv.VerifyReceipt.ResponseBody").
		Info(body)

	var resp apple.VerificationResponseBody
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logger.WithField("trace", "IAPEnv.VerifyReceipt").Error(err)
		return resp, err
	}

	return resp, nil
}

// When a receipt is verified, or webhook is triggered,
// we dissects the data
// and records those data derived from it:
//
// FailureSchema
// SessionSchema
// only specific to verification.
//
// and WebHookSchema specific to notification.
//
// And all should record those data:
// Transaction array
// PendingRenewal array
// ReceiptToken
// Subscription

// SaveVerificationSession saves the root fields
// of VerificationResponseBody.
func (env IAPEnv) SaveVerificationSession(v apple.VerificationSessionSchema) error {
	_, err := env.db.NamedExec(insertVerificationSession, v)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveVerificationSession").Error(err)
		return err
	}

	return nil
}

// Save transaction save an entry of user's transaction
// history.
// UnifiedReceipt.LatestTransactions field.
func (env IAPEnv) SaveTransaction(r apple.TransactionSchema) error {
	_, err := env.db.NamedExec(insertTransaction, r)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveTransaction").Error(err)
		return err
	}

	return nil
}

// SavePendingRenewal saves the UnifiedReceipt.PendingRenewalInfo array.
func (env IAPEnv) SavePendingRenewal(p apple.PendingRenewalSchema) error {
	_, err := env.db.NamedExec(insertPendingRenewal, p)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SavePendingRenewal").Error(err)
		return err
	}

	return nil
}

// CreateSubscription saves an Subscription instance
// built from the latest transaction.
func (env IAPEnv) CreateSubscription(s apple.Subscription) error {
	_, err := env.db.NamedExec(insertIAPSubscription, s)

	if err != nil {
		logger.WithField("trace", "IAPEnv.CreateSubscription").Error(err)
		return err
	}

	return nil
}

// SaveReceiptTokenDB saves the base-64 encoded receipt data
// for one original transaction id.
func (env IAPEnv) SaveReceiptTokenDB(r apple.ReceiptToken) error {
	_, err := env.db.NamedExec(insertReceiptToken, r)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveReceiptTokenDB").Error(err)

		return err
	}

	return nil
}
