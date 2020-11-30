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
