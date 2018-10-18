package model

import "gitlab.com/ftchinese/subscription-api/util"

// Membership contains a user's membership details
type Membership struct {
	UserID string
	Tier   string
	Cycle  string
	Start  string
	Expire string
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
	var vipType int64
	var expireTime int64

	err := env.DB.QueryRow(query, userID).Scan(
		&m.UserID,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.Start,
		&m.Expire,
	)

	if err != nil {
		logger.WithField("location", "Retrieve membership").Error(err)

		return m, err
	}

	if m.Tier == "" {
		m.Tier = normalizeMemberTier(vipType)
	}

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

// NewMember creates a new member
func (env Env) NewMember(m Membership) error {
	query := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		start_utc = ?,
		expire_utc = ?`

	_, err := env.DB.Exec(query,
		m.UserID,
		m.Tier,
		m.Cycle,
		m.Start,
		m.Expire,
	)

	if err != nil {
		logger.WithField("location", "Create a new member").Error(err)

		return err
	}

	return nil
}

// RenewMember renews membership duration before piror to expiration.
func (env Env) RenewMember(m Membership) error {
	query := `
	UPDATE premium.ftc_vip
	SET member_tier = ?,
		billing_cyce = ?,
		expire_utc = ?
	WHERE vip_id = ?
	LIMIT 1`

	_, err := env.DB.Exec(query,
		m.Tier,
		m.Cycle,
		m.Expire,
		m.UserID,
	)

	if err != nil {
		logger.WithField("location", "Renew membership").Error(err)

		return err
	}

	return nil
}
