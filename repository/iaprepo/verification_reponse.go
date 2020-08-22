package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
)

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

// SaveVerifiedReceipt saves the root fields
// of VerificationResp.
func (env Env) SaveVerifiedReceipt(v apple.VerifiedReceiptSchema) error {
	_, err := env.db.NamedExec(apple.StmtSaveVerifiedReceipt, v)

	if err != nil {
		logger.WithField("trace", "Env.SaveVerifiedReceipt").Error(err)
		return err
	}

	return nil
}

// Save transaction save an entry of user's transaction history.
// UnifiedReceipt.LatestTransactions field.
func (env Env) SaveTransaction(r apple.TransactionSchema) error {
	_, err := env.db.NamedExec(apple.StmtInsertTransaction, r)

	if err != nil {
		logger.WithField("trace", "Env.SaveTransaction").Error(err)
		return err
	}

	return nil
}

// SavePendingRenewal saves the UnifiedReceipt.PendingRenewalInfo array.
func (env Env) SavePendingRenewal(p apple.PendingRenewalSchema) error {
	_, err := env.db.NamedExec(apple.StmtSavePendingRenewal, p)

	if err != nil {
		logger.WithField("trace", "Env.SavePendingRenewal").Error(err)
		return err
	}

	return nil
}

// SaveReceiptTokenDB saves the base-64 encoded receipt data
// for one original transaction id.
func (env Env) SaveReceiptTokenDB(r apple.ReceiptToken) error {
	_, err := env.db.NamedExec(apple.StmtSaveReceiptToken, r)

	if err != nil {
		logger.WithField("trace", "Env.SaveReceiptTokenDB").Error(err)

		return err
	}

	return nil
}
