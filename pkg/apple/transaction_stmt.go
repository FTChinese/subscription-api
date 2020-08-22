package apple

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
quantity = :quantity
`

// StmtInsertTransaction saves an item in UnifiedReceipt.LatestTransactions
const StmtInsertTransaction = `
INSERT IGNORE INTO premium.apple_transaction
SET ` + transactionBaseCols + `,
	web_order_line_item_id = :web_order_line_item_id,
	cancellation_date_ms = :cancellation_date_ms,
	cancellation_reason = :cancellation_reason,
	is_upgraded = :is_upgraded,
	promotional_offer_id = :promotional_offer_id,
	subscription_group_id = :sub_group_id,
	created_utc = UTC_TIMESTAMP()`
