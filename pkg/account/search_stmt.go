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

const StmtSearchByEmail = `
SELECT user_id AS id
FROM cmstmp01.userinfo
WHERE email = ?
LIMIT 1`

// StmtSearchByMobile tries to find a row from the profile
// db by mobile_phone column.
const StmtSearchByMobile = `
SELECT profile.user_id AS id
FROM user_db.profile
WHERE mobile_phone = ?
LIMIT 1`
