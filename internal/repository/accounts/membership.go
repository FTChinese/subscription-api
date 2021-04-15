package accounts

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

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
