package wxlogin

import "github.com/guregu/null"

// AccountStatus is used to check if an FTC account is bound to wechat account,
// and the account's membership.
type AccountStatus struct {
	UnionID       null.String // The union id bound to this FTC account. If valid, it means the user_id is bound to wx_union_id.
	IsMember      bool        // Is this FTC account a subscribed member?
	MemberUnionID null.String // For backward compatibility. Membership might already bound in ftc_vip table.
	ExpireDate    null.String // If this FTC account is a paid member, when will it expire?
}

// CheckFTCAccount finds out if an ftc account exists.
// If exists, whether this ftc account is bound to a wechat account,
// whether the bound wechat account is the target one,
// whether is ftc account has a membership.
func (env Env) CheckFTCAccount(userID string) (AccountStatus, error) {
	// For this query, if UnionID is valid, it means the ftc account is already bound to a wechat account.
	// Previous to this API taking effect, in premium.ftc_vip table, many rows already have both ftc user id and wechat union id set.
	// In such cases, both CheckFTCAccount and CheckWxAccount will retrieve the save membership but the userinfo.wx_union_id is empty.
	// If the right side of LEFT JOIN is empty, memberUnionId will be null;
	// If vip_id_alias is not set, it is an empty string;
	// If vip_id_alias is set, it is non-empty string.
	query := `
	SELECT u.wx_union_id AS unionId,
		v.vip_id IS NOT NULL AS isMember,
		NULLIF(v.vip_id_alias, '') AS memberUnionId,
		IFNULL(v.expire_time, 0) AS expireTime,
		v.expire_date AS expireDate
	FROM cmstmp01.userinfo AS u
		LEFT JOIN premium.ftc_vip AS v
		ON u.user_id = v.vip_id
	WHERE u.user_id = ?
	LIMIT 1`

	var s AccountStatus
	var expireTime int64

	err := env.DB.QueryRow(query, userID, userID).Scan(
		&s.UnionID,
		&s.IsMember,
		&s.MemberUnionID,
		&expireTime,
		&s.ExpireDate,
	)

	if err != nil {
		logger.WithField("trace", "CheckFTCAccount").Error(err)
		return s, err
	}

	if !s.ExpireDate.Valid {
		expireDate := normalizeExpireDate(expireTime)
		if expireDate != "" {
			s.ExpireDate = null.StringFrom(expireDate)
		}
	}

	return s, nil
}

// CheckWxAccount finds out if a wechat account exists.
// If exists, whether this wechat account has a membership.
func (env Env) CheckWxAccount(unionID string) (AccountStatus, error) {
	// For this query, if membership exists,
	// memberUserId must be valid, and memberUnionId must be valid.
	// Here unionId == memberUnionId == unionID
	query := `
	SELECT w.unionid AS unionId,
		v.vip_id IS NOT NULL AS isMember,
		NULLIF(v.vip_id_alias, '') AS memberUnionId,
		IFNULL(v.expire_time, 0) AS expireTime,
		v.expire_date AS expireDate
	FROM user_db.user_sns_info AS w
		LEFT JOIN premium.ftc_vip AS v
		ON w.unionid = v.vip_id_alias
	WHERE w.unionid = ?
	LIMIT 1`

	var s AccountStatus
	var expireTime int64

	err := env.DB.QueryRow(query, unionID, unionID).Scan(
		&s.UnionID,
		&s.IsMember,
		&s.MemberUnionID,
		&expireTime,
		&s.ExpireDate,
	)

	if err != nil {
		logger.WithField("trace", "CheckWxAccount").Error(err)
		return s, err
	}

	if !s.ExpireDate.Valid {
		expireDate := normalizeExpireDate(expireTime)
		if expireDate != "" {
			s.ExpireDate = null.StringFrom(expireDate)
		}
	}

	return s, nil
}
