package iaprepo

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

const insertWebHook = `
INSERT IGNORE INTO premium.apple_webhook
SET ` + transactionBaseCols + `,
	app_item_id = :app_item_id,
	item_id = :item_id,
	auto_renew_adam_id = :auto_renew_adam_id,
	auto_renew_product_id = :auto_renew_product_id,
	auto_renew_status = :auto_renew_status,
	auto_renew_status_change_date_ms = :auto_renew_status_change_date_ms,
	expiration_intent = :expiration_intent,
	notification_type = :notification_type,
	encryption_password = :password,
	response_status = :status,
	created_utc = UTC_TIMESTAMP()`
