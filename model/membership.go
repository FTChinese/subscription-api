package model

import (
	"database/sql"
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

// CanRenew tests if current member is allowed to renuew subscription.
// A member could only renew its subscripiton one billing cycle ahead of current cycle.
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

// IsExpired tests is membership is expired.
func (m Membership) IsExpired() bool {
	t, err := time.Parse(time.RFC3339, m.Expire)

	if err != nil {
		return true
	}

	// If expire is before now, it is expired.
	return t.Before(time.Now())
}

// NewMemberFromOrder create a membership based the an order.
// ConfirmedAt field must be empty at this step.
func (env Env) NewMemberFromOrder(order SubscribeOrder, confirmTime time.Time) Membership {
	expireTime := order.CalculateExpireTime(confirmTime)

	startAt := util.SQLDatetimeUTC.FromTime(confirmTime)
	expireAt := util.SQLDatetimeUTC.FromTime(expireTime)

	member, err := env.Membership(order.UserID)

	if err != nil && err == sql.ErrNoRows {
		member.UserID = order.UserID
		member.Tier = order.TierToBuy
		member.Cycle = order.BillingCycle
		member.Start = startAt
		member.Expire = expireAt

		return member
	}

	// If membership existed but expired,
	if member.IsExpired() {
		member.Tier = order.TierToBuy
		member.Cycle = order.BillingCycle
		member.Start = startAt
		member.Expire = expireAt

		return member
	}

	// Membership renewal. Just extend expiration time.
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

// CreateMember creates a new member or renew it.
// A member might already exits but expired, and now he is re-subscribe.
// In such case the ON DUPLICATE KEY UPDATE clause will take effect.
func (env Env) CreateMember(m Membership) error {
	query := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		start_utc = ?,
		expire_utc = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		start_utc = ?,
		expire_utc = ?`

	_, err := env.DB.Exec(query,
		m.UserID,
		string(m.Tier),
		string(m.Cycle),
		m.Start,
		m.Expire,
		string(m.Tier),
		string(m.Cycle),
		m.Start,
		m.Expire,
	)

	if err != nil {
		logger.WithField("location", "Create a new member").Error(err)

		return err
	}

	return nil
}
