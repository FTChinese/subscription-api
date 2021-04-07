package account

const StmtIDExists = `
SELECT EXISTS (
	SELECT *
	FROM cmstmp01.userinfo
	WHERE user_id = ?)`

const StmtEmailExists = `
SELECT EXISTS (
	SELECT *
	FROM cmstmp01.userinfo
	WHERE email = ?)`

const StmtNameExists = `
SELECT EXISTS (
	SELECT *
	FROM cmstmp01.userinfo
	WHERE user_name = ?)`
