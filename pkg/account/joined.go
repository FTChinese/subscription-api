package account

// JoinedSchema is used to retrieve
type JoinedSchema struct {
	BaseAccount
	Wechat
	VIP bool `db:"is_vip"`
}

// When retrieve joined account, use union_id from wechat_userinfo table.
const colsJoinedAccount = colsBaseAccount + `,
IFNULL(u.is_vip, FALSE) 		AS is_vip,
w.union_id 						AS wx_union_id,
w.nickname AS wx_nickname,
w.avatar_url AS wx_avatar_url
`

// StmtJoinedAccountByFtc selects ftc + wechat account
// by ftc uuid.
const StmtJoinedAccountByFtc = colsJoinedAccount + `
FROM cmstmp01.userinfo AS u
LEFT JOIN user_db.profile AS p
	ON u.user_id = p.user_id
LEFT JOIN user_db.wechat_userinfo AS w
	ON u.wx_union_id = w.union_id
WHERE u.user_id = ?
LIMIT 1`

// StmtJoinedAccountByWx select ftc + wechat from wechat side.
const StmtJoinedAccountByWx = colsJoinedAccount + `
FROM user_db.wechat_userinfo AS w
LEFT JOIN cmstmp01.userinfo AS u
	ON w.union_id = u.wx_union_id
LEFT JOIN user_db.profile AS p
	ON u.user_id = p.user_id
WHERE w.union_id = ?
LIMIT 1`
