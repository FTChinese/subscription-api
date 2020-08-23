package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// BackUpMember takes a snapshot of membership.
func (env Env) BackUpMember(snapshot reader.MemberSnapshot) error {

	_, err := env.db.NamedExec(
		reader.StmtSnapshotMember(env.cfg.GetSubsDB()),
		snapshot)

	if err != nil {
		logger.WithField("trace", "Env.BackUpMember").Error(err)

		return err
	}

	return nil
}
