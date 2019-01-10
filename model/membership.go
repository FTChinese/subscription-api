package model

import (
	"fmt"
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Membership contains a user's membership details
type Membership struct {
	UserID     string      `json:"-"`
	UnionID    null.String `json:"-"`
	Tier       enum.Tier   `json:"tier"`
	Cycle      enum.Cycle  `json:"billingCycle"`
	ExpireDate util.Date   `json:"expireDate"` // On which date the membership ends
}

// canRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renwal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
func (m Membership) canRenew(cycle enum.Cycle) bool {
	cycleEnds, err := cycle.EndingTime(time.Now())

	if err != nil {
		return false
	}

	return m.ExpireDate.Before(cycleEnds)
}

// isExpired tests if the membership's expiration date is before now.
func (m Membership) isExpired() bool {
	// If expire is before now, it is expired.
	return m.ExpireDate.Before(time.Now())
}

// findMember retrieves a user's membership based on subscription information.
func (env Env) findMember(subs Subscription) (Membership, error) {
	var whereCol string

	if subs.isWxLogin() {
		whereCol = "vip_id_alias"
	} else {
		whereCol = "vip_id"
	}

	query := fmt.Sprintf(`
	SELECT vip_id AS userId,
		vip_id_alias AS unionId,
		vip_type AS vipType,
		expire_time AS expireTime,
		member_tier AS memberTier,
		billing_cycle AS billingCyce,
		expire_date AS expireDate
	FROM premium.ftc_vip
	WHERE %s = ?
	LIMIT 1`, whereCol)

	var m Membership
	var vipType int64
	var expireTime int64

	err := env.DB.QueryRow(query, subs.UserID).Scan(
		&m.UserID,
		&m.UnionID,
		&vipType,
		&expireTime,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate,
	)

	if err != nil {
		logger.WithField("trace", "findMember").Error(err)

		return m, err
	}

	if !m.Tier.IsValid() {
		m.Tier = normalizeMemberTier(vipType)
	}

	if !m.Cycle.IsValid() {
		m.Cycle = enum.CycleYear
	}

	if m.ExpireDate.IsZero() {
		m.ExpireDate = util.DateFrom(normalizeExpireDate(expireTime))
	}

	return m, nil
}
