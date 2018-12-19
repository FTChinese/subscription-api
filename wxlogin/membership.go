package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
)

// Membership contains user's subscription data.
// If membership is not bound to another account yet,
// UserID should not be empty if membership is purcahsed with FTC account while UnionID must be null;
// UserID should not be empty and UnionID should not be null, and their values are equal if membership is purchased with Wechat account;
// UserID should not be empty and UnionID should not be null, and their values should be different if FTC account is already bound to Wechat account.
// If membership is not bound to another account yet, either UserID or UnionID should be empty, if membership actually present.
type Membership struct {
	UserID     string      `json:"-"`
	UnionID    null.String `json:"-"`
	Tier       enum.Tier   `json:"tier"`
	Cycle      enum.Cycle  `json:"billingCycle"`
	ExpireDate string      `json:"expireDate"`
	ExpireTime time.Time   `json:"-"`
}

// IsExpired test if a membership is expired.
func (m Membership) IsExpired() bool {
	if m.ExpireTime.IsZero() {
		return true
	}

	return m.ExpireTime.Before(time.Now())
}

// IsEqualTo tests if two memberships are the same one.
// If return false, it indicates the two accounts are bound to 3rd accounts, or not bound to any account.
func (m Membership) IsEqualTo(other Membership) bool {
	return m.UserID == other.UserID
}

// IsCoupled tests if a membership is bound to another account.
// Any membership that is coupled to another one should deny merge request.
func (m Membership) IsCoupled() bool {
	return m.UnionID.Valid && (m.UnionID.String != m.UserID)
}

// IsFromFTC tests if this membership is purchased by ftc account
func (m Membership) IsFromFTC() bool {
	return !m.UnionID.Valid
}

// IsFromWx tests if this membership is purchased by wechat account
func (m Membership) IsFromWx() bool {
	return m.UnionID.Valid && (m.UnionID.String == m.UserID)
}

// IsEmpty tests if a membership is empty.
// Empty membership is defined as UserID == "" since the ftc_vip.vip_id column must have a value.
func (m Membership) IsEmpty() bool {
	return m.UserID == ""
}

// Merge merges wechat membership into ftc membership and returns a new Membership.
// Both m and other should not be an empty Membership.
func (m Membership) Merge(other Membership) Membership {
	merged := Membership{}

	if m.IsFromFTC() {
		merged.UserID = m.UserID
	} else if other.IsFromFTC() {
		merged.UserID = m.UserID
	}

	if m.IsFromWx() {
		merged.UnionID = m.UnionID
	} else if other.IsFromWx() {
		merged.UnionID = other.UnionID
	}

	if m.Tier > other.Tier {
		merged.Tier = m.Tier
	} else {
		merged.Tier = other.Tier
	}

	if m.Cycle > other.Cycle {
		merged.Cycle = m.Cycle
	} else {
		merged.Cycle = other.Cycle
	}

	if m.ExpireTime.After(other.ExpireTime) {
		merged.ExpireTime = m.ExpireTime
	} else {
		merged.ExpireTime = other.ExpireTime
	}

	return merged
}

// Pick picks from two membership the one that is not empty.
func (m Membership) Pick(other Membership) Membership {
	if m.IsEmpty() && other.IsEmpty() {
		return m
	}

	if !m.IsEmpty() && !other.IsEmpty() {
		return m.Merge(other)
	}

	if !m.IsEmpty() {
		return m
	}

	return other
}

// func (env Env) FindMemberByFTC(userID string, c chan Membership) error {
// 	query := `
// 	SELECT vip_id AS userId,
// 		vip_id_alias AS unionId,
// 		vip_type AS vipType,
// 		expire_time AS expireTime,
// 		member_tier AS memberTier,
// 		billing_cycle AS billingCycle,
// 		IFNULL(expire_date, '') AS expireDate
// 	FROM premium.ftc_vip
// 	WHERE vip_id = ?
// 	LIMIT 1`

// 	var m Membership
// 	var vipType int64
// 	var expireTime int64

// 	err := env.DB.QueryRow(query, userID).Scan(
// 		&m.UserID,
// 		&m.UnionID,
// 		&vipType,
// 		&expireTime,
// 		&m.Tier,
// 		&m.Cycle,
// 		&m.ExpireDate,
// 	)

// 	if err != nil {
// 		logger.Error(err)
// 		c <- m
// 		return err
// 	}

// 	if !m.Tier.IsValid() {
// 		m.Tier = normalizeMemberTier(vipType)
// 	}

// 	if m.ExpireDate == "" {
// 		m.ExpireDate = normalizeExpireDate(expireTime)
// 	}

// 	m.ExpireTime, err = util.ParseDateTime(m.ExpireDate, time.UTC)

// 	if err != nil {
// 		c <- m
// 		return err
// 	}

// 	c <- m
// 	return nil
// }
