package apple

const StmtSavePendingRenewal = `
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
