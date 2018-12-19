package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Wechat is a concise version of UserInfo,
// containing only essential data to identify a wechat user.
type Wechat struct {
	UnionID   string `json:"unionId"`
	OpenID    string `json:"openId"`
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

// Account is a user's FTC account.
// If ID is empty, it means the Wechat account is not bound to FTC account.
type Account struct {
	UserID     string      `json:"id"` // Will be empty is a wechat account is not bound to an ftc account.
	UserName   null.String `json:"userName"`
	Email      string      `json:"email"`
	AvatarURL  null.String `json:"avatarUrl"`
	IsVIP      bool        `json:"isVip"`
	IsVerified bool        `json:"isVerified"`
	Wechat     *Wechat     `json:"wechat"` // will be nil if ftc account is not bound to a wechat account
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
	return (a.UserID != "") && (a.Wechat != nil)
}

// IsMember checks if an account is a paid member.
func (a Account) IsMember() bool {
	return !a.Membership.IsEmpty()
}

// FindAccountByWx retrieves a user account by wechat union id.
// Wechat account is essential piece in this SQL statement while FTC account and membership are optional.
func (env Env) FindAccountByWx(unionID string) (Account, error) {
	query := `
	SELECT w.unionid AS unionId,
		w.openid AS openId,
		w.nickname AS nickName,
		w.headimgurl AS avatarUrl,
		IFNULL(v.vip_id, ''),
		v.vip_id_alias,
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
	FROM user_db.user_sns_info AS w
		LEFT JOIN premium.ftc_vip AS v
		ON w.unionid = v.vip_id_alias
		LEFT JOIN cmstmp01.userinfo AS u
		ON w.unionid = u.wx_union_id
	WHERE w.unionid = ?
	LIMIT 1`

	var a Account
	var wx Wechat
	var vipType int64
	var expireTime int64
	var m Membership

	err := env.DB.QueryRow(query, unionID).Scan(
		&wx.UnionID,
		&wx.OpenID,
		&wx.NickName,
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
	a.Wechat = &wx
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
		u.isvip AS isVip,
		u.active AS isVerified,
		IFNULL(v.vip_id, ''),
		v.vip_id_alias,
		IFNULL(v.vip_type, 0) AS vipType,
		IFNULL(v.expire_time, 0) AS expireTime,
		member_tier AS memberTier,
		v.billing_cycle AS billingCyce,
		IFNULL(v.expire_date, '') AS expireDate,
		IFNULL(unionid, '') AS unionId,
		IFNULL(openid, '') AS openId,
		IFNULL(nickname, '') AS nickName,
		IFNULL(headimgurl, '') AS avatarUrl
	FROM cmstmp01.userinfo AS u
		LEFT JOIN premium.ftc_vip AS v
		ON u.user_id = v.vip_id
		LEFT JOIN user_db.user_sns_info AS w
		ON u.wx_union_id = w.unionid
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
		&wx.OpenID,
		&wx.NickName,
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

	if wx.UnionID != "" {
		a.Wechat = &wx
	}

	a.Membership = m

	return a, nil
}

// func (env Env) FindWxAccount(unionID string, c chan Wechat) error {
// 	query := `
// 	SELECT unionid AS unionId,
// 		openid AS openid,
// 		nickname AS nickName,
// 		headimgurl AS avatarUrl
// 	FROM user_db.user_sns_info
// 	WHERE unionid = ?
// 	LIMIT 1`

// 	var w Wechat
// 	err := env.DB.QueryRow(query, unionID).Scan(
// 		&w.UnionID,
// 		&w.OpenID,
// 		&w.NickName,
// 		&w.AvatarURL,
// 	)

// 	if err != nil {
// 		logger.Error(err)
// 		c <- w
// 		return err
// 	}

// 	c <- w
// 	return nil
// }
