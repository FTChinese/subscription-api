package wxlogin

import (
	"gitlab.com/ftchinese/subscription-api/model"
)

// WxAccount is a concise version of UserInfo,
// containing only essential data to identify a wechat user.
type WxAccount struct {
	UnionID   string `json:"unionId"`
	OpenID    string `json:"openId"`
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

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

// SubscribeMethod indicates what kind of account user used when subscribe to membership.
type SubscribeMethod int

// Login methods when user subscribed to membership.
const (
	MethodNone  SubscribeMethod = 0
	MethodEmail SubscribeMethod = 1
	MethodWx    SubscribeMethod = 2
)

// BoundAccount tells which acounts should be bound,
// and how to merge them.
type BoundAccount struct {
	UserID  string
	UnionID string
	Method  SubscribeMethod
}

// BindAccount binds an ftc account to wechat account.
func (env Env) BindAccount(userID, unionID string) error {
	query := `UPDATE cmstmp01.userinfo
	SET wx_union_id = ?
	WHERE user_id = ?
	LIMIT 1`

	_, err := env.DB.Exec(query, unionID, userID)

	if err != nil {
		logger.WithField("trace", "BindAccount").Error(err)
		return err
	}

	return nil
}

// MergeMembership merges from membership either from ftc to wechat, or vice versus.
func (env Env) MergeMembership(b BoundAccount) error {
	var err error
	switch b.Method {
	// If user logged in with ftc account and subscribed to mebership, set vip_id_alias column to wechat union id
	case MethodEmail:
		stmtMergeMember := `
		UPDATE premium.ftc_vip
		SET vip_id_alias = ?
		WHERE vip_id = ?
		LIMIT 1`
		_, err = env.DB.Exec(stmtMergeMember, b.UnionID, b.UserID)

	case MethodWx:
		stmtMergeMember := `
		UPDATE premium.ftc_vip
		SET vip_id = ?
		WHERE vip_id_alias = ?
		LIMIT 1`

		_, err = env.DB.Exec(stmtMergeMember, b.UserID, b.UnionID)
	}

	if err != nil {
		logger.WithField("trace", "MergeMemership").Error(err)
		return err
	}

	return nil
}

// MergeAccount associate a wechat account with an FTC account.
// The FTC account must not be bound to a wechat account,
// And must not subscribed to any kind of membership.
// It set the wx_union_id column to wechat unioin id and set the membership's vip_id column to user id.
func (env Env) MergeAccount(b BoundAccount) error {
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

	_, errA := tx.Exec(stmtUnionID, b.UnionID, b.UserID)

	// Error 1062: Duplicate entry 'ogfvwjk6bFqv2yQpOrac0J3PqA0o' for key 'wx_union_id'
	if errA != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount set union id").Error(errA)
	}

	var errB error
	// If user subscribed with wechat,
	// replace vip_id with ftc user id.
	// If user subscribed with ftc account,
	// replace vip_id_alis with wechat union id.
	switch b.Method {
	// If user logged in with ftc account and subscribed to mebership, set vip_id_alias column to wechat union id
	case MethodEmail:
		stmtMergeMember := `
		UPDATE premium.ftc_vip
		SET vip_id_alias = ?
		WHERE vip_id = ?
		LIMIT 1`
		_, errB = tx.Exec(stmtMergeMember, b.UnionID, b.UserID)

	case MethodWx:
		stmtMergeMember := `
		UPDATE premium.ftc_vip
		SET vip_id = ?
		WHERE vip_id_alias = ?
		LIMIT 1`

		_, errB = tx.Exec(stmtMergeMember, b.UserID, b.UnionID)
	}

	// Error 1062: Duplicate entry 'e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae' for key 'PRIMARY'"
	// If the `userID` is already a member.
	if errB != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount").Error(errB)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "BindAccount commit trasaction").Error(err)

		return err
	}

	return nil
}

// LoadAccountByWx retrieves a user's wechat account with membership
func (env Env) LoadAccountByWx(unionID string) (Account, error) {
	query := `
	SELECT w.unionid AS unionId,
		w.openid AS openId,
		w.nickname AS nickName,
		w.headimgurl AS avatarUrl,
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
	FROM user_db.user_sns_info AS w
		LEFT JOIN premium.ftc_vip AS v
		ON w.unionid = v.vip_id_alias
		LEFT JOIN cmstmp01.userinfo AS u
		ON w.unionid = u.wx_union_id
	WHERE w.unionid = ?
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
		&a.Email,
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

	a.Wechat = &wx
	a.Membership = m

	return a, nil
}
