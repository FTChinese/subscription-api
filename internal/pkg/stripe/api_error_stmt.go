package stripe

const StmtSaveAPIError = `
INSERT INTO premium.log_stripe_error
SET ftc_user_id = ?,
	charge_id = ?,
	error_code = ?,
	http_status = ?,
	error_message = ?,
	parameter = ?,
	request_id = ?,
	error_type = ?,
	created_utc = UTC_TIMESTAMP()`
