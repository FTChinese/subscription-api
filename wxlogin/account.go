package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Wechat contains a user's Wechat account data.
type Wechat struct {
	UnionID   string `json:"unionId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

// exists tests if Wechat account exists.
func (w Wechat) exists() bool {
	return w.UnionID != ""
}

// Account is a user's FTC account.
// If ID is empty, it means the Wechat account is not bound to FTC account.
// Wechat field is alwasy not null for a user account retrieved from
// subscription-api.
type Account struct {
	UserID     string      `json:"id"` // Will be empty is a wechat account is not bound to an ftc account.
	UserName   null.String `json:"userName"`
	Email      string      `json:"email"`
	AvatarURL  null.String `json:"avatarUrl"`
	IsVIP      bool        `json:"isVip"`
	IsVerified bool        `json:"isVerified"`
	Wechat     Wechat      `json:"wechat"` // will be nil if ftc account is not bound to a wechat account
	Membership Membership  `json:"membership"`
}

// IsEqualTo tests if account retrieve by ftc user id is the same one as retrieved by wechat union id.
// If true, it means the two accounts are the same one -- this ftc account is already bound to the wechat account.
// If false, it could only indicates the two accounts are not bound:
// 1. FTC account might bind to another wechat account;
// 2. Wechat account might bind to another FTC account;
// 3. They are bind to none.
func (a Account) IsEqualTo(wx Account) bool {
	return a.UserID == wx.UserID
}

// IsCoupled tests if an ftc account is bound to a wechat account, or vice versus.
// If an account is bound to another one, its UserID field must not be empty, and Wechat must not be nil.
// Futhermore, its UserID must not be equal to Wechat.UnionID.
func (a Account) IsCoupled() bool {
	return (a.UserID != "") && a.Wechat.exists()
}

// IsMember checks if an account is a paid member.
func (a Account) IsMember() bool {
	return !a.Membership.IsEmpty()
}

// FindAccountByWx retrieves a user account by wechat union id.
// Wechat account is essential piece in this SQL statement while FTC account and membership are optional.
func (env Env) FindAccountByWx(unionID string) (Account, error) {
	query := `
	SELECT w.union_id AS unionId,
		w.nickname AS nickname,
		w.avatar_utl AS avatarUrl,
		IFNULL(v.vip_id, '') AS mUserID,
		v.vip_id_alias AS mUnionId,
		IFNULL(v.vip_type, 0) AS vipType,
		IFNULL(v.expire_time, 0) AS expireTime,
		v.member_tier AS memberTier,
		v.billing_cycle AS billingCyce,
		IFNULL(v.expire_date, '') AS expireDate,
		IFNULL(u.user_id, '') AS id,
		u.user_name AS userName,
		IFNULL(u.email, '') AS email,
		IFNULL(u.isvip, 0) AS isVip,
		IFNULL(u.active, 0) AS isVerified
	FROM user_db.wechat_userinfo AS w
		LEFT JOIN premium.ftc_vip AS v
		ON w.union_id = v.vip_id_alias
		LEFT JOIN cmstmp01.userinfo AS u
		ON w.union_id = u.wx_union_id
	WHERE w.union_id = ?
	LIMIT 1`

	var a Account
	var wx Wechat
	var vipType int64
	var expireTime int64
	var m Membership

	err := env.DB.QueryRow(query, unionID).Scan(
		&wx.UnionID,
		&wx.Nickname,
		&wx.AvatarURL,
		&m.UserID,
		&m.UnionID,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate,
		&a.UserID,
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

	if m.ExpireDate == "" {
		m.ExpireDate = normalizeExpireDate(expireTime)
	}

	if m.ExpireDate != "" {
		m.ExpireTime, _ = util.ParseDateTime(m.ExpireDate, time.UTC)
	}

	// wx is always not nil
	a.Wechat = wx
	a.Membership = m

	return a, nil
}

// FindAccountByFTC retreives a user account by ftc user id.
// Although data retrieves is similar to LoadAccountByWx, the core point is quite different.
// This uses userinfo table as essential requirements while membership and wechat account are optional.
func (env Env) FindAccountByFTC(userID string) (Account, error) {
	query := `
	SELECT u.user_id AS id,
		u.user_name AS userName,
		u.email AS email,
		u.is_vip AS isVip,
		u.email_verified AS isVerified,
		IFNULL(v.vip_id, '') AS mUserId,
		v.vip_id_alias AS mUnionId,
		IFNULL(v.vip_type, 0) AS vipType,
		IFNULL(v.expire_time, 0) AS expireTime,
		member_tier AS memberTier,
		v.billing_cycle AS billingCyce,
		IFNULL(v.expire_date, '') AS expireDate,
		IFNULL(w.union_id, '') AS unionId,
		IFNULL(w.nickname, '') AS nickname,
		IFNULL(w.avatar_url, '') AS avatarUrl
	FROM cmstmp01.userinfo AS u
		LEFT JOIN premium.ftc_vip AS v
		ON u.user_id = v.vip_id
		LEFT JOIN user_db.wechat_userinfo AS w
		ON u.wx_union_id = w.union_id
	WHERE u.user_id = ?
	LIMIT 1`

	var a Account
	var wx Wechat
	var vipType int64
	var expireTime int64
	var m Membership

	err := env.DB.QueryRow(query, userID).Scan(
		&a.UserID,
		&a.UserName,
		&a.Email,
		&a.IsVIP,
		&a.IsVerified,
		&m.UserID,
		&m.UnionID,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate,
		&wx.UnionID,
		&wx.Nickname,
		&wx.AvatarURL,
	)

	if err != nil {
		logger.WithField("trace", "FindAccountByFTC").Error(err)

		return a, err
	}

	if !m.Tier.IsValid() {
		m.Tier = normalizeMemberTier(vipType)
	}

	if m.ExpireDate == "" {
		m.ExpireDate = normalizeExpireDate(expireTime)
	}

	if m.ExpireDate != "" {
		m.ExpireTime, _ = util.ParseDateTime(m.ExpireDate, time.UTC)
	}

	a.Wechat = wx
	a.Membership = m

	return a, nil
}
