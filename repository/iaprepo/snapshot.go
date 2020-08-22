package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// BackUpMember takes a snapshot of membership.
func (env Env) BackUpMember(snapshot subs.MemberSnapshot) error {

	_, err := env.db.NamedExec(
		subs.StmtSnapshotMember(env.cfg.GetSubsDB()),
		snapshot)

	if err != nil {
		logger.WithField("trace", "Env.BackUpMember").Error(err)

		return err
	}

	return nil
}
