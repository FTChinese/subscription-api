package apple

const StmtSaveVerifiedReceipt = `
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
