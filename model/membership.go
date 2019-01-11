package model

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Duration contains a membership's expiration time.
// This type exits for compatibility due to expiration time are saved into two columns.
type Duration struct {
	timestamp  int64
	ExpireDate util.Date `json:"expireDate"`
}

func (d *Duration) normalizeDate() {
	if d.ExpireDate.IsZero() && d.timestamp != 0 {
		d.ExpireDate = util.DateFrom(time.Unix(d.timestamp, 0))
	}
}

// canRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renwal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
func (d Duration) canRenew(cycle enum.Cycle) bool {
	cycleEnds, err := cycle.EndingTime(time.Now())

	if err != nil {
		return false
	}

	return d.ExpireDate.Before(cycleEnds)
}

// isExpired tests if the membership's expiration date is before now.
func (d Duration) isExpired() bool {
	// If expire is before now, it is expired.
	return d.ExpireDate.Before(time.Now())
}

// Membership contains a user's membership details
type Membership struct {
	UserID   string      `json:"-"`
	UnionID  null.String `json:"-"`
	Tier     enum.Tier   `json:"tier"`
	Cycle    enum.Cycle  `json:"billingCycle"`
	Duration             // On which date the membership ends
}

// findMember retrieves a user's membership based on subscription information.
func (env Env) findMember(subs Subscription) (Membership, error) {
	query := subs.stmtMember()

	var m Membership
	var vipType int64

	err := env.DB.QueryRow(query, subs.UserID).Scan(
		&m.UserID,
		&m.UnionID,
		&vipType,
		&m.Tier,
		&m.Cycle,
		&m.timestamp,
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

	m.normalizeDate()

	return m, nil
}
