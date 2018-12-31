package model

import (
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Membership contains a user's membership details
type Membership struct {
	UserID string     `json:"-"`
	Tier   enum.Tier  `json:"tier"`
	Cycle  enum.Cycle `json:"billingCycle"`
	Expire string     `json:"expireDate"` // On which date the membership ends
}

// CanRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renwal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
func (m Membership) CanRenew(cycle enum.Cycle) bool {
	expireDate, err := util.ParseSQLDate(m.Expire)

	if err != nil {
		logger.WithField("location", "Parse expiration date")
		return false
	}

	afterACycle, err := cycle.TimeAfterACycle(time.Now())

	if err != nil {
		return false
	}

	return expireDate.Before(afterACycle)
}

// IsExpired tests if the membership saved in database is expired.
func (m Membership) IsExpired() bool {
	t, err := util.ParseSQLDate(m.Expire)

	if err != nil {
		return true
	}

	// If expire is before now, it is expired.
	return t.Before(time.Now())
}

// FindMember retrieves a user's membership
func (env Env) FindMember(userID string) (Membership, error) {
	query := `
	SELECT vip_id AS userId,
		vip_type AS vipType,
		expire_time AS expireTime,
		IFNULL(member_tier, '') AS memberTier,
		IFNULL(billing_cycle, '') AS billingCyce,
		IFNULL(expire_date, '') AS expireDate
	FROM premium.ftc_vip
	WHERE vip_id = ?
	LIMIT 1`

	var m Membership
	var tier string
	var cycle string
	var vipType int64
	var expireTime int64

	err := env.DB.QueryRow(query, userID).Scan(
		&m.UserID,
		&vipType,
		&expireTime,
		&tier,
		&cycle,
		&m.Expire,
	)

	if err != nil {
		logger.WithField("location", "FindMember").Error(err)

		return m, err
	}

	if tier == "" {
		m.Tier = normalizeMemberTier(vipType)
	} else {
		m.Tier = enum.NewTier(tier)
	}

	m.Cycle, _ = enum.NewCycle(cycle)

	if m.Expire == "" {
		m.Expire = normalizeExpireDate(expireTime)
	}

	return m, nil
}
