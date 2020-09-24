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
// ReceiptToken
// Subscription

// SaveDecodedReceipt saved the decoded apple.ClientReceipt from response body.
func (env Env) SaveDecodedReceipt(v apple.VerifiedReceiptSchema) error {
	_, err := env.db.NamedExec(apple.StmtSaveDecodedReceipt, v)

	if err != nil {
		return err
	}

	return nil
}

// SaveTransaction save an entry of user's transaction history.
// UnifiedReceipt.LatestReceiptInfo field.
// Duplicate entries are ignored.
func (env Env) SaveTransaction(r apple.TransactionSchema) error {
	_, err := env.db.NamedExec(apple.StmtInsertTransaction, r)

	if err != nil {
		return err
	}

	return nil
}

// SavePendingRenewal saves an item in the the UnifiedReceipt.PendingRenewalInfo array.
// Duplicate entries are ignores.
func (env Env) SavePendingRenewal(p apple.PendingRenewalSchema) error {
	_, err := env.db.NamedExec(apple.StmtSavePendingRenewal, p)

	if err != nil {
		return err
	}

	return nil
}

// SaveReceiptTokenDB saves the base-64 encoded receipt data
// for one original transaction id.
// It is not in use since we changed to disk files.
// TODO: try to save it in HBase.
func (env Env) SaveReceiptTokenDB(r apple.ReceiptToken) error {
	_, err := env.db.NamedExec(apple.StmtSaveReceiptToken, r)

	if err != nil {
		return err
	}

	return nil
}

// SaveResponsePayload saves the payload of common fields
// in either verification response or webhook. It includes:
//
// * LatestReceiptInfo
// * PendingRenewalInfo
// * LatestReceipt
//
// The LatestReceipt field is saved on disk as a file for now.
// This part is separated since it its also used in webhook.
func (env Env) SaveResponsePayload(ur apple.UnifiedReceipt) {
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

	rt := ur.ReceiptToken()

	// Save the LatestReceipt field to a file.
	// Initially it was designed to save in SQL.
	sugar.Info("Saving receipt to disk")
	go func() {
		_ = SaveReceiptTokenFile(rt)
	}()

	// Save in redis
	sugar.Infof("Saving receipt to redis")
	go func() {
		err := env.rdb.Set(
			ctx,
			rt.ReceiptKeyName(),
			rt.LatestReceipt,
			0,
		).Err()
		if err != nil {
			sugar.Errorf("error saving receipt to redis: %s", err)
		}
	}()
}
