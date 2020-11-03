package apple

const colShared = `
environment = :environment,
original_transaction_id = :original_transaction_id
`

// StmtSaveDecodedReceipt saves the decoded receipt.
// It includes all the fields of `receipt` field in a
// verification response.
const StmtSaveDecodedReceipt = `
INSERT IGNORE INTO premium.apple_verification_session
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
	transaction_id = :transaction_id,
	app_item_id = :app_item_id,
	application_version = :application_version,
	bundle_id = :bundle_id,
	download_id = :download_id,
	expiration_date_ms = :expiration_date_ms,
	original_purchase_date_ms = :original_purchase_date_ms,
	original_application_version = :original_application_version,
	preorder_date_ms = :preorder_date_ms,
	receipt_creation_date_ms = :receipt_creation_date_ms,
	receipt_type = :receipt_type,
	request_date_ms = :request_date_ms,
	response_status = :status,
	version_external_identifier = :version_external_identifier,
	created_utc = UTC_TIMESTAMP()`

// Corresponding statement for BaseTransactionSchema.
const transactionBaseCols = `
environment = :environment,
original_transaction_id = :original_transaction_id,
transaction_id = :transaction_id,
expires_date_ms = :expires_date_ms,
is_in_intro_offer_period = :is_in_intro_offer_period,
is_trial_period = :is_trial_period,
original_purchase_date_ms = :original_purchase_date_ms,
product_id = :product_id,
purchase_date_ms = :purchase_date_ms,
quantity = :quantity,
web_order_line_item_id = :web_order_line_item_id
`

// StmtInsertTransaction is SQL fro TransactionSchema that saves an item in UnifiedReceipt.LatestReceiptInfo
const StmtInsertTransaction = `
INSERT IGNORE INTO premium.apple_transaction
SET ` + transactionBaseCols + `,
	cancellation_date_ms = :cancellation_date_ms,
	cancellation_reason = :cancellation_reason,
	is_upgraded = :is_upgraded,
	promotional_offer_id = :promotional_offer_id,
	subscription_group_id = :sub_group_id,
	created_utc = UTC_TIMESTAMP()`

const colsPendingRenewal = `
auto_renew_product_id = :auto_renew_product_id,
auto_renew_status = :auto_renew_status,
expiration_intent = :expiration_intent,
grace_period_expires_date_ms = :grace_period_expires_date_ms,
is_in_billing_retry_period = :is_in_billing_retry_period,
price_consent_status = :price_consent_status,
product_id = :product_id
`

// StmtSavePendingRenewal is SQL for PendingRenewalSchema that is used to save the items in UnifiedReceipt.PendingRenewalInfo.
const StmtSavePendingRenewal = `
INSERT IGNORE INTO premium.apple_pending_renewal
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
` + colsPendingRenewal + `,
	updated_utc = UTC_TIMESTAMP(),
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + colsPendingRenewal + `,
	updated_utc = UTC_TIMESTAMP()`
