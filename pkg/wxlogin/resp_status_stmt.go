package wxlogin

const StmtSaveWxRespError = `
INSERT INTO user_db.wechat_error_log
SET code = :code,
	message = :message,
	created_utc = UTC_TIMESTAMP()`
