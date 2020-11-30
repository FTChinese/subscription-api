package readerrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) RetrieveMember(id reader.MemberID) (reader.Membership, error) {
	var m reader.Membership

	err := env.db.Get(
		&m,
		reader.StmtSelectMember,
		id.BuildFindInSet())

	if err != nil && err != sql.ErrNoRows {
		return reader.Membership{}, err
	}

	return m.Normalize(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (env Env) RetrieveAppleMember(txID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.db.Get(
		&m,
		reader.StmtAppleMember,
		txID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Normalize(), nil
}
