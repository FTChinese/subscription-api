package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// BackUpMember saves a member's snapshot at a specific moment.
func (env SubEnv) BackUpMember(snapshot subs.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		subs.StmtSnapshotMember(env.GetSubsDB()),
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

// FindBalanceSources finds all orders with unused portion.
// This is identical to OrderTx.FindBalanceSources without a transaction.
func (env SubEnv) FindBalanceSources(id reader.MemberID) ([]subs.ProratedOrderSchema, error) {
	var sources = make([]subs.ProratedOrderSchema, 0)

	err := env.db.Select(
		&sources,
		subs.StmtBalanceSource(env.GetSubsDB()),
		id.CompoundID,
		id.UnionID)

	if err != nil {
		logger.WithField("trace", "SubEnv.FindBalanceSources").Error(err)
		return sources, err
	}

	return sources, nil
}
