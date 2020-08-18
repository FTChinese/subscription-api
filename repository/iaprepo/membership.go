package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// BackUpMember takes a snapshot of membership.
func (env IAPEnv) BackUpMember(snapshot subs.MemberSnapshot) error {

	_, err := env.db.NamedExec(
		subs.StmtSnapshotMember(env.c.GetSubsDB()),
		snapshot)

	if err != nil {
		logger.WithField("trace", "IAPEnv.BackUpMember").Error(err)

		return err
	}

	return nil
}
