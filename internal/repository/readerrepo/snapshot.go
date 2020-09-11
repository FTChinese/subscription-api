package readerrepo

import "github.com/FTChinese/subscription-api/pkg/reader"

// BackUpMember saves a member's snapshot at a specific moment.
func (env Env) BackUpMember(snapshot reader.MemberSnapshot) error {
	_, err := env.db.NamedExec(
		reader.StmtSnapshotMember,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}
