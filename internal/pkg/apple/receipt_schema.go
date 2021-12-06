package apple

import "github.com/FTChinese/go-rest/chrono"

// Save the receipt as a token for status polling.
const StmtSaveReceiptToken = `
INSERT IGNORE INTO file_store.apple_receipt
SET original_transaction_id = :original_transaction_id,
	environment = :environment,
	receipt = :receipt,
	created_utc = UTC_TIMESTAMP()`

const StmtRetrieveReceipt = `
SELECT receipt
FROM file_store.apple_receipt
WHERE original_transaction_id = ?
	AND environment = ?
LIMIT 1`

// ReceiptSchema is a row to save the receipt data from a request
// so that later we can re-validate this receipt.
type ReceiptSchema struct {
	BaseSchema
	LatestReceipt string      `db:"receipt"`
	CreatedUTC    chrono.Time `db:"created_utc"`
}

func NewReceiptSchema(u UnifiedReceipt) ReceiptSchema {
	return ReceiptSchema{
		BaseSchema: BaseSchema{
			Environment:           u.Environment,
			OriginalTransactionID: u.latestTransaction.OriginalTransactionID,
		},
		LatestReceipt: u.LatestReceipt,
		CreatedUTC:    chrono.TimeNow(),
	}
}
