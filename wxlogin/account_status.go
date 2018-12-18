package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// FTCAccountStatus is used to inspect a FTC account status:
// whether it has a wechat account bound;
// whether it is a paid user;
// whether its membership is bound to a wechat account;
// whether, if bound account and membership is bound to a wechat account, the bounded wechat account is the same one.
type FTCAccountStatus struct {
	UnionID       null.String // The union id bound to this FTC account. If valid, it means the user_id is bound to wx_union_id.
	IsMember      bool        // Is this FTC account a subscribed member?
	MemberUnionID null.String // For backward compatibility. Membership might already bound in ftc_vip table.
	ExpireDate    null.String // If this FTC account is a paid member, when will it expire?
	IsExpired     bool
}

// IsBound test if ftc account has a wechat account bound.
func (s FTCAccountStatus) IsBound() bool {
	return s.UnionID.Valid
}

// IsBoundTo test if this ftc account is bound to the specified wechat account, in case the ftc account is already bound.
// If the ftc account is already bound to the wechat account, noop.
// If the ftc account is not bound to this unionID. forbid further operation.
func (s FTCAccountStatus) IsBoundTo(unionID string) bool {
	return s.UnionID.String == unionID
}

// IsEqualTo tests if a the membership of FTCAccountStatus is bound to the same as wechat account as the one belonging to WxAccountStatus.
// If the columns user_id and wx_union_id in userinfo are already bound, it should return true;
// if it returns false, it means the same user's ftc account and wechat have separate membership!
// This is a serious problem that must be addressed manually.
func (s FTCAccountStatus) IsEqualTo(wx WxAccountStatus) bool {
	return s.MemberUnionID.String == wx.UnionID
}

// CheckFTCAccount finds out if an ftc account exists.
// If exists, whether this ftc account is bound to a wechat account,
// whether the bound wechat account is the target one,
// whether is ftc account has a membership.
func (env Env) CheckFTCAccount(userID string) (FTCAccountStatus, error) {
	// For this query, if UnionID is valid, it means the ftc account is already bound to a wechat account.
	// Previous to this API taking effect, in premium.ftc_vip table, many rows already have both ftc user id and wechat union id set.
	// In such cases, both CheckFTCAccount and CheckWxAccount will retrieve the save membership but the userinfo.wx_union_id is empty.
	// If the right side of LEFT JOIN is empty, memberUnionId will be null;
	// If v.vip_id IS NOT NULL while v.vip_id_alias is empty, it means the membership is not bound.
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

	var s FTCAccountStatus
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

	if !s.ExpireDate.Valid {
		s.IsExpired = true
	} else {
		t, err := util.ParseSQLDate(s.ExpireDate.String)
		if err != nil {
			s.IsExpired = true
		}

		s.IsExpired = t.Before(time.Now())
	}

	return s, nil
}

// WxAccountStatus is used to to get account data logged in via wechat.
type WxAccountStatus struct {
	UnionID    string
	IsMember   bool        // Is this FTC account a subscribed member? Test the vip_id column to get this value.
	ExpireDate null.String // If this FTC account is a paid member, when will it expire?
	IsExpired  bool
}

// CheckWxAccount finds out if a wechat account exists.
// If exists, whether this wechat account has a membership.
func (env Env) CheckWxAccount(unionID string) (WxAccountStatus, error) {
	// For this query, if membership exists,
	// memberUserId must be valid, and memberUnionId must be valid.
	// For a LEFT JOIN, w.unionid should always be NOT NULL, while v.vip_id might not exist.
	// If v.vip_id IS NOT NULL, then v.vip_id_alias must be NOT NULL.
	query := `
	SELECT w.unionid AS unionId,
		v.vip_id IS NOT NULL AS isMember,
		v.vip_id AS memberUserId,
		IFNULL(v.expire_time, 0) AS expireTime,
		v.expire_date AS expireDate
	FROM user_db.user_sns_info AS w
		LEFT JOIN premium.ftc_vip AS v
		ON w.unionid = v.vip_id_alias
	WHERE w.unionid = ?
	LIMIT 1`

	var s WxAccountStatus
	var expireTime int64

	err := env.DB.QueryRow(query, unionID, unionID).Scan(
		&s.UnionID,
		&s.IsMember,
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

	if !s.ExpireDate.Valid {
		s.IsExpired = true
	} else {
		t, err := util.ParseSQLDate(s.ExpireDate.String)
		if err != nil {
			s.IsExpired = true
		}

		s.IsExpired = t.Before(time.Now())
	}

	return s, nil
}

// func (env Env) FindAccountByFTC(userID string, c chan FTCAccount) error {
// 	query := `
// 	SELECT user_id AS id,
// 		user_name AS userName,
// 		email AS email,
// 		isvip AS isVip,
// 		active AS isVerified
// 	FROM cmstmp01.userinfo
// 	WHERE user_id = ?
// 	LIMIT 1`

// 	var a FTCAccount
// 	err := env.DB.QueryRow(query, userID).Scan(
// 		&a.UserID,
// 		&a.UnionID,
// 		&a.UserName,
// 		&a.Email,
// 		&a.IsVIP,
// 		&a.IsVerified,
// 	)

// 	if err != nil {
// 		logger.Error(err)
// 		c <- a
// 		return err
// 	}

// 	c <- a
// 	return nil
// }
