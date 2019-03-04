package model

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// findMember retrieves a user's membership based on subscription information.
func (env Env) findMember(subs paywall.Subscription) (paywall.Membership, error) {
	query := env.stmtSelectMember()

	var m paywall.Membership
	var vipType int64

	err := env.db.QueryRow(
		query,
		subs.CompoundID,
		subs.UnionID,
	).Scan(
		&m.CompoundID,
		&m.UnionID,
		&vipType,
		&m.Tier,
		&m.Cycle,
		&m.Timestamp,
		&m.ExpireDate,
	)

	if err != nil {
		logger.WithField("trace", "findMember").Error(err)

		return m, err
	}

	if m.Tier == enum.InvalidTier {
		m.Tier = normalizeMemberTier(vipType)
	}

	if m.Cycle == enum.InvalidCycle {
		m.Cycle = enum.CycleYear
	}

	m.NormalizeDate()

	return m, nil
}
