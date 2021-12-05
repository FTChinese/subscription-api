package apple

const StmtLoggingWebhook = `
INSERT IGNORE INTO premium.apple_webhook
SET ` + colShared + `,
	auto_renew_adam_id = :auto_renew_adam_id,
	auto_renew_product_id = :auto_renew_product_id,
	auto_renew_status = :auto_renew_status,
	auto_renew_status_change_date_ms = :auto_renew_status_change_date_ms,
	expiration_intent = :expiration_intent,
	notification_type = :notification_type,
	encryption_password = :password,
	response_status = :status,
	created_utc = UTC_TIMESTAMP()`
