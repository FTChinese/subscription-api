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

// buildMembership create a membership based on an subscripiton.
// ConfirmedAt field must be empty at this step.
// func (env Env) buildMembership(s Subscription, confirmTime time.Time) Membership {
// 	// Calculate expiration time based the when this subscription is confirmed and the billing cycle.
// 	expireTime := s.DeduceExpireTime(confirmTime)
// 	logger.WithField("location", "buidlMembership").Infof("Decuced exppire time based on subsbription cofirmation time: %s", expireTime)

// 	expireDate := util.SQLDateUTC.FromTime(expireTime)

// 	// Try to find out if this subscription order's owner is already a member, or used to be a member.
// 	member, err := env.FindMember(s.UserID)

// 	// If there's any error (including sql.ErrNoRows), create a new mebership.
// 	// Err here should not prevent user becoming a member.
// 	if err != nil {
// 		logger.WithField("location", "Build membership").Infof("Membership for user %s not found. Assuming this is a new member", s.UserID)

// 		member.UserID = s.UserID
// 		member.Tier = s.TierToBuy
// 		member.Cycle = s.BillingCycle
// 		member.Expire = expireDate

// 		return member
// 	}

// 	// Membership exists. See if curent membership is expired. If expired, update startDate and expireDate.
// 	if member.IsExpired() {
// 		logger.WithField("location", "Build membership").Infof("Membership for user %s found but expired.", s.UserID)

// 		member.Tier = s.TierToBuy
// 		member.Cycle = s.BillingCycle
// 		member.Expire = expireDate

// 		return member
// 	}

// 	// Membership exists, and it is not expired.
// 	// It means user is renewing subscription.
// 	// Just extend the expiration time.
// 	logger.WithField("location", "Build membership").Infof("Membership for user %s found for renewal. Will expire: %s", s.UserID, member.Expire)

// 	member.Expire = s.RenewExpireDate(member.Expire)

// 	logger.WithField("location", "buildMembership").Infof("Membership extended: %s", member.Expire)

// 	return member
// }

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
		m.Expire = normalizeExpireDate(expireTime)
	}

	return m, nil
}
