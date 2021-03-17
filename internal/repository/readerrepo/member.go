package readerrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) RetrieveMember(id pkg.UserIDs) (reader.Membership, error) {
	var m reader.Membership

	err := env.dbs.Read.Get(
		&m,
		reader.StmtSelectMember,
		id.BuildFindInSet())

	if err != nil && err != sql.ErrNoRows {
		return reader.Membership{}, err
	}

	return m.Sync(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (env Env) RetrieveAppleMember(txID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.dbs.Read.Get(
		&m,
		reader.StmtAppleMember,
		txID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

func (env Env) UpdateMember(m reader.Membership) error {
	_, err := env.dbs.Write.NamedExec(
		reader.StmtUpdateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}
