package subs

const StmtSaveConfirmResult = `
INSERT INTO premium.confirmation_result
SET order_id = :order_id,
	succeeded = :succeeded,
	failed = :failed,
	created_utc = UTC_TIMESTAMP()`
