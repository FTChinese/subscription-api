package iaprepo

const insertVerificationSession = `
INSERT IGNORE INTO premium.apple_verification_session
SET response_status = :status,
	environment = :environment,
	receipt_type = :receipt_type,
	app_item_id = :app_item_id,
	bundle_id = :bundle_id,
	application_version = :application_version,
	download_id = :download_id,
	version_external_identifier = :version_external_identifier,
	receipt_creation_date_ms = :receipt_creation_date_ms,
	original_purchase_date_ms = :original_purchase_date_ms,
	original_application_version = :original_application_version,
	expiration_date_ms = :expiration_date_ms,
	preorder_date_ms = :preorder_date_ms,
	created_utc = UTC_TIMESTAMP()`

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

const insertTransaction = `
INSERT IGNORE INTO premium.apple_transaction
SET ` + transactionBaseCols + `,
	web_order_line_item_id = :web_order_line_item_id,
	cancellation_date_ms = :cancellation_date_ms,
	cancellation_reason = :cancellation_reason,
	is_upgraded = :is_upgraded,
	promotional_offer_id = :promotional_offer_id,
	subscription_group_id = :sub_group_id,
	created_utc = UTC_TIMESTAMP()`

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

const insertPendingRenewal = `
INSERT IGNORE INTO premium.apple_pending_renewal
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
	product_id = :product_id,
	auto_renew_status = :auto_renew_status,
	expiration_intent = :expiration_intent,
	auto_renew_product_id = :auto_renew_product_id,
	is_in_billing_retry_period = :is_in_billing_retry_period,
	grace_period_expires_date_ms = :grace_period_expires_date_ms,
	price_consent_status = :price_consent_status,
	updated_utc = UTC_TIMESTAMP(),
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
	auto_renew_status = :auto_renew_status,
	expiration_intent = :expiration_intent,
	auto_renew_product_id = :auto_renew_product_id,
	is_in_billing_retry_period = :is_in_billing_retry_period,
	grace_period_expires_date_ms = :grace_period_expires_date_ms,
	price_consent_status = :price_consent_status,
	updated_utc = UTC_TIMESTAMP()`

const iapSubBaseCols = `
last_transaction_id = :last_transaction_id,
product_id = :product_id,
purchase_date_utc = :purchase_date_utc,
expires_date_utc = :expires_date_utc,
tier = :plan_tier,
cycle = :plan_cycle,
auto_renewal = :auto_renewal,
updated_utc = UTC_TIMESTAMP()
`

const insertIAPSubscription = `
INSERT INTO premium.apple_subscription
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
` + iapSubBaseCols + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + iapSubBaseCols

// Save the receipt as a token for status polling.
const insertReceiptToken = `
INSERT INTO premium.apple_receipt_token
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
	latest_receipt = :latest_receipt,
	updated_utc = UTC_TIMESTAMP(),
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
	latest_receipt = :latest_receipt,
	updated_utc = UTC_TIMESTAMP()`
