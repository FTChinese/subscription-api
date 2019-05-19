package model

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// RetrieveMember retrieves a user's membership info by ftc id
// or wechat union id.
func (env Env) RetrieveMember(u paywall.User) (paywall.Membership, error) {

	var m paywall.Membership

	// In the ftc_vip table, vip_id might be ftc uuid or wechat
	// id, and vip_id_alias is always wechat id.
	// In future, the table will be refactor with two extra
	// columns ftc_user_id dedicated to ftc uuid and wx_union_id
	// dedicated for wechat union id. The vip_id column will be
	// use only as a unique constraint on these two columns.
	err := env.db.QueryRow(
		env.query.SelectMember(),
		u.CompoundID,
		u.UnionID,
	).Scan(
		&m.CompoundID,
		&m.UnionID,
		&m.Tier,
		&m.Cycle,
		&m.ExpireDate)

	if err != nil {
		logger.WithField("trace", "RetrieveMember").Error(err)

		return m, err
	}

	// TODO: this is not accurate.
	if m.Cycle == enum.InvalidCycle {
		m.Cycle = enum.CycleYear
	}

	return m, nil
}
