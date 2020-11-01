package wxlogin

const colUserInfo = `
nickname = :nickname,
avatar_url = :avatar_url,
gender = :gender,
country = :country,
province = :province,
city = :city,
privilege = :privilege,
updated_utc = UTC_TIMESTAMP()
`

const StmtInsertUserInfo = `
INSERT INTO user_db.wechat_userinfo
SET union_id = :union_id,
` + colUserInfo + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + colUserInfo

const StmtUpdateUserInfo = `
UPDATE user_db.wechat_userinfo
SET ` + colUserInfo + `
WHERE union_id = :union_id`
