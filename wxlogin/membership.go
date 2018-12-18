package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
)

// Membership contains user's subscription data.
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
func (m Membership) IsEqualTo(other Membership) bool {
	return m.UserID == other.UserID
}

// Merge merges wechat membership into ftc membership and returns a new Membership.
func (m Membership) Merge(wx Membership) Membership {
	merged := Membership{
		UserID:  m.UserID,
		UnionID: wx.UnionID,
	}

	if m.Tier > wx.Tier {
		merged.Tier = m.Tier
	} else {
		merged.Tier = wx.Tier
	}

	if m.Cycle > wx.Cycle {
		merged.Cycle = m.Cycle
	} else {
		merged.Cycle = wx.Cycle
	}

	if m.ExpireTime.After(wx.ExpireTime) {
		merged.ExpireTime = m.ExpireTime
	} else {
		merged.ExpireTime = wx.ExpireTime
	}

	return merged
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
