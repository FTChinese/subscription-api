package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
)

// When a receipt is verified, or webhook is triggered,
// we dissects the data
// and records those data derived from it:
//
// FailureSchema
// ReceiptSchema
// only specific to verification.
//
// and WebHookSchema specific to notification.
//
// And all should record those data:
// Transaction array
// PendingRenewal array
// ReceiptSchema
// Subscription

// SaveDecodedReceipt saved the decoded apple.ClientReceipt from response body.
func (env Env) SaveDecodedReceipt(v apple.VerifiedReceiptSchema) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtSaveDecodedReceipt, v)

	if err != nil {
		return err
	}

	return nil
}

// SaveTransaction save an entry of user's transaction history.
// UnifiedReceipt.LatestReceiptInfo field.
// Duplicate entries are ignored.
func (env Env) SaveTransaction(r apple.TransactionSchema) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtInsertTransaction, r)

	if err != nil {
		return err
	}

	return nil
}

// SavePendingRenewal saves an item in the the UnifiedReceipt.PendingRenewalInfo array.
// Duplicate entries are ignores.
func (env Env) SavePendingRenewal(p apple.PendingRenewalSchema) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtSavePendingRenewal, p)

	if err != nil {
		return err
	}

	return nil
}

// SaveUnifiedReceipt saves the payload of common fields
// in either verification response or webhook. It includes:
//
// * LatestReceiptInfo
// * PendingRenewalInfo
// * LatestReceipt
//
// The LatestReceipt field is saved on disk as a file for now.
// This part is separated since it its also used in webhook.
func (env Env) SaveUnifiedReceipt(ur apple.UnifiedReceipt) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	// Save the LatestReceiptInfo array.
	sugar.Info("Saving receipt history array")
	go func() {
		for _, v := range ur.LatestReceiptInfo {
			_ = env.SaveTransaction(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save PendingRenewalInfo array
	sugar.Info("Saving pending renewal array")
	go func() {
		for _, v := range ur.PendingRenewalInfo {
			_ = env.SavePendingRenewal(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save receipt file.
	go func() {
		env.SaveReceipt(apple.NewReceiptSchema(ur))
	}()
}
