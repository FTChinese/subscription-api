package readerrepo

import "github.com/FTChinese/subscription-api/pkg/reader"

// ArchiveMember saves a member's snapshot at a specific moment.
func (env Env) ArchiveMember(snapshot reader.MemberSnapshot) error {
	_, err := env.dbs.Write.NamedExec(
		reader.StmtSnapshotMember,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}
