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
	Start  string
	Expire string
}

// CanRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
func (m Membership) CanRenew(cycle BillingCycle) bool {
	expireAt, err := time.Parse(time.RFC3339, m.Expire)

	if err != nil {
		logger.WithField("location", "Parse expiration time")
		return false
	}

	now := time.Now()

	switch cycle {
	case Yearly:
		// expiration time < today + cycle
		return expireAt.Before(now.AddDate(1, 0, 0))
	case Monthly:
		return expireAt.Before(now.AddDate(0, 1, 0))
	}

	return false
}

// IsExpired tests is the membership saved in database is expired.
func (m Membership) IsExpired() bool {
	t, err := time.Parse(time.RFC3339, m.Expire)

	if err != nil {
		return true
	}

	// If expire is before now, it is expired.
	return t.Before(time.Now())
}

// NewMemberFromOrder create a membership based on an order.
// ConfirmedAt field must be empty at this step.
func (env Env) NewMemberFromOrder(s Subscription, confirmTime time.Time) Membership {
	// Calculate expiration time based the when this subscription is confirmed and the billing cycle.
	expireTime := s.DeduceExpireTime(confirmTime)

	// Use confirmTime as membership's starting point.
	// Convert it into SQL DATETIME string in UTC.
	startAt := util.SQLDatetimeUTC.FromTime(confirmTime)
	// Convert expiration time into SQL DATETIME string in UTC.
	expireAt := util.SQLDatetimeUTC.FromTime(expireTime)

	// Try to find out if this subscription order's owner is already a member, or used to be a member.
	member, err := env.Membership(s.UserID)

	// If there's any error (including sql.ErrNoRows), create a new mebership.
	if err != nil {
		member.UserID = s.UserID
		member.Tier = s.TierToBuy
		member.Cycle = s.BillingCycle
		member.Start = startAt
		member.Expire = expireAt

		return member
	}

	// Membership exists. If it is expired, treat is a new subscription.
	if member.IsExpired() {
		member.Tier = s.TierToBuy
		member.Cycle = s.BillingCycle
		member.Start = startAt
		member.Expire = expireAt

		return member
	}

	// Membership exists, and it is not expired.
	// It means user is renewing subscription.
	// Just extend the expiration time.
	member.Expire = expireAt

	return member
}

// Membership retrieves a user's membership
func (env Env) Membership(userID string) (Membership, error) {
	query := `
	SELECT vip_id AS userId,
		vip_type AS vipType,
		expire_time AS expireTime,
		IFNULL(v.member_tier, '') AS memberTier,
		IFNULL(v.billing_cycle, '') AS billingCyce,
		IFNULL(v.start_utc, '') AS startAt,
		IFNULL(v.expire_utc, '') AS expireAt
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
		&m.Start,
		&m.Expire,
	)

	if err != nil {
		logger.WithField("location", "Retrieve membership").Error(err)

		return m, err
	}

	if tier == "" {
		m.Tier = normalizeMemberTier(vipType)
	} else {
		m.Tier, _ = NewTier(tier)
	}

	m.Cycle, _ = NewCycle(cycle)

	if m.Expire == "" {
		m.Expire = normalizeExpireTime(expireTime)
	} else {
		// Convert UTC DATETIME to UTC ISO8601
		m.Expire = util.ISO8601UTC.FromDatetime(m.Expire, nil)
	}

	if m.Start == "" {
		m.Start = normalizeStartTime(expireTime)
	} else {
		m.Start = util.ISO8601UTC.FromDatetime(m.Start, nil)
	}

	return m, nil
}
