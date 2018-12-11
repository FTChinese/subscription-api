package wxlogin

import "gitlab.com/ftchinese/subscription-api/model"

// WxAccount is user's wx identity.
type WxAccount struct {
	OpenID    string `json:"openId"`
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
	UnionID   string `json:"unionId"`
}

// Account is a user's account
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

// FindAccount tries to see if a user's wechat account is bound to an email.
func (env Env) FindAccount(unionID string) (Account, error) {
	query := `SELECT u.user_id AS id,
		IFNULL(user_name, '') AS userName,
		email AS email
		isvip AS isVip,
		active AS isVerified
	FROM cmstmp01.userinfo
	WHERE wx_union_id = ?
	LIMIT 1`

	var a Account
	err := env.DB.QueryRow(query, unionID).Scan(
		&a.ID,
		&a.UserName,
		&a.Email,
		&a.IsVIP,
		&a.IsVerified,
	)

	// sql.ErrNoRows
	if err != nil {
		logger.WithField("trace", "FindAccount").Error(err)

		return a, err
	}

	return a, nil
}

// FindMembership tries to find membership by wechat union id.
func (env Env) FindMembership(unionID string) (model.Membership, error) {
	query := `SELECT vip_id AS userId,
		vip_type AS vipType,
		expire_time AS expireTime,
		member_tier AS memberTier,
		billing_cycle AS billingCycle,
		IFNULL(expire_date, '') AS expireDate
	FROM premium.ftc_vip
	WHERE vip_id_alias = ?
	LIMIT 1`

	var m model.Membership
	var vipType int64
	var expireTime int64
	err := env.DB.QueryRow(query, unionID).Scan(
		&m.UserID,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.Expire,
	)

	if err != nil {
		logger.WithField("trace", "FindMembership").Error(err)
		return m, err
	}

	if !m.Tier.IsValid() {
		m.Tier = normalizeMemberTier(vipType)
	}

	if m.Expire == "" {
		m.Expire = normalizeExpireDate(expireTime)
	}

	return m, nil
}
