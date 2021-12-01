package shared

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// RetrieveMember loads reader.Membership of the specified id.
// compoundID - Might be ftc uuid or wechat union id.
func (env ReaderBaseRepo) RetrieveMember(compoundID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtSelectMember,
		compoundID)

	if err != nil && err != sql.ErrNoRows {
		return reader.Membership{}, err
	}

	return m.Sync(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (env ReaderBaseRepo) RetrieveAppleMember(txID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtAppleMember,
		txID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

// ArchiveMember saves a member's snapshot at a specific moment.
// Deprecated.
func (env ReaderBaseRepo) ArchiveMember(snapshot reader.MemberSnapshot) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtSaveSnapshot,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

func (env ReaderBaseRepo) VersionMembership(v reader.MembershipVersioned) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtVersionMembership,
		v)

	if err != nil {
		return err
	}

	return nil
}
