package wxlogin

// colUserInfo maps to the shared fields of userinfo.
// Open id is ignored in this table since the same user
// might have multiple open id.
const colUserInfo = `
avatar_url = :avatar_url,
city = :city,
country = :country,
gender = :gender,
nickname = :nickname,
province = :province,
privilege = :privilege,
updated_utc = :updated_utc
`

const StmtUpsertUserInfo = `
INSERT INTO user_db.wechat_userinfo
SET union_id = :union_id,
` + colUserInfo + `,
	created_utc = :created_utc
ON DUPLICATE KEY UPDATE
` + colUserInfo

// Deprecated
const StmtUpdateUserInfo = `
UPDATE user_db.wechat_userinfo
SET ` + colUserInfo + `
WHERE union_id = :union_id`
