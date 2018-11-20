package model

import (
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

// Membership contains a user's membership details
type Membership struct {
	UserID string
	Tier   MemberTier
	Cycle  BillingCycle
	Expire string // On which date the membership ends
}

// CanRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
func (m Membership) CanRenew(cycle BillingCycle) bool {
	expireDate, err := util.ParseSQLDate(m.Expire)

	if err != nil {
		logger.WithField("location", "Parse expiration date")
		return false
	}

	now := time.Now()

	// Add one day more to accomodate timezone change.
	switch cycle {
	case Yearly:
		// expiration time < today + cycle
		return expireDate.Before(now.AddDate(1, 0, 1))
	case Monthly:
		return expireDate.Before(now.AddDate(0, 1, 1))
	}

	return false
}

// IsExpired tests is the membership saved in database is expired.
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
		m.Tier, _ = NewTier(tier)
	}

	m.Cycle, _ = NewCycle(cycle)

	if m.Expire == "" {
		m.Expire = normalizeExpireDate(expireTime)
	}

	return m, nil
}
