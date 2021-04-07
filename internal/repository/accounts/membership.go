package accounts

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

type memberAsyncResult struct {
	value reader.Membership
	err   error
}

func (env Env) AsyncMembership(ids pkg.UserIDs) <-chan memberAsyncResult {
	c := make(chan memberAsyncResult)

	go func() {
		m, err := env.RetrieveMember(ids)

		c <- memberAsyncResult{
			value: m,
			err:   err,
		}
	}()

	return c
}
