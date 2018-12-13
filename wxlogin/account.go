package wxlogin

import "gitlab.com/ftchinese/subscription-api/model"

// Account is a user's FTC account.
// If ID is empty, it means the Wechat account is not bound to FTC account.
type Account struct {
	ID         string           `json:"id"`
	UserName   string           `json:"userName"`
	Email      string           `json:"email"`
	AvatarURL  string           `json:"avatarUrl"`
	IsVIP      bool             `json:"isVip"`
	IsVerified bool             `json:"isVerified"`
	Wechat     *WxAccount       `json:"wechat"`
	Membership model.Membership `json:"membership"`
}

// BindAccount associate a wechat account with an FTC account.
// It set the wx_union_id column to wechat unioin id and set the membership's vip_id column to user id.
func (env Env) BindAccount(userID, unionID string) error {
	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("trace", "BindAccount").Error(err)
		return err
	}

	// Update the wx_union_id field of a user's account based on user id.
	stmtUnionID := `
	UPDATE cmstmp01.userinfo
	SET wx_union_id = ?
	WHERE user_id = ?
	LIMIT 1`

	_, errA := tx.Exec(stmtUnionID, userID)

	if errA != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount set union id").Error(err)
	}

	// Set the premium.ftc_vip table's vip_id columnd to user id.
	stmtMemberID := `
	UPDATE premium.ftc_vip
	SET vip_id = ?
	WHERE vip_id_alias = ?
	LIMIT 1`

	_, errB := tx.Exec(stmtMemberID, unionID)

	if errB != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount set membership user id").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "BindAccount commit trasaction").Error(err)

		return err
	}

	return nil
}

// FindAccountByWx retrieves a user's wechat account with membership
func (env Env) FindAccountByWx(unionID string) (Account, error) {
	query := `
	SELECT unionid AS unionId,
		openid AS openId,
		nickname AS nickName,
		headimgurl AS avatarUrl,
		IFNULL(v.vip_type, 0) AS vipType,
		IFNULL(v.expire_time, 0) AS expireTime,
		v.member_tier AS memberTier,
		v.billing_cycle AS billingCyce,
		IFNULL(v.expire_date, '') AS expireDate,
		IFNULL(u.user_id, '') AS id,
		IFNULL(u.user_name, '') AS userName,
		IFNULL(u.email, '') AS email,
		IFNULL(u.isvip, 0) AS isVip,
		IFNULL(u.active, 0) AS isVerified
	FROM user_db.user_sns_info AS i
		LEFT JOIN premium.ftc_vip AS v
		ON i.unionid = v.vip_id_alias
		LEFT JOIN cmstmp01.userinfo AS u
		ON i.unionid = u.wx_union_id
	WHERE unionid = ?
	LIMIT 1`

	var a Account
	var wx WxAccount
	var vipType int64
	var expireTime int64
	var m model.Membership

	err := env.DB.QueryRow(query, unionID).Scan(
		&wx.UnionID,
		&wx.OpenID,
		&wx.NickName,
		&wx.AvatarURL,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.Expire,
		&a.ID,
		&a.UserName,
		&a.IsVIP,
		&a.IsVerified,
	)

	if err != nil {
		logger.WithField("trace", "FindBoundAccount").Error(err)

		return a, err
	}

	if !m.Tier.IsValid() {
		m.Tier = normalizeMemberTier(vipType)
	}

	if m.Expire == "" {
		m.Expire = normalizeExpireDate(expireTime)
	}

	return a, nil
}
