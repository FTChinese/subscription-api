package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) JoinedByFtcID(ftcID string) (account.JoinedSchema, error) {
	var s account.JoinedSchema
	err := env.dbs.Read.Get(&s, account.StmtJoinedAccountByFtc, ftcID)
	if err != nil {
		return account.JoinedSchema{}, err
	}

	return s, nil
}

func (env Env) JoinedByWxID(unionID string) (account.JoinedSchema, error) {
	var s account.JoinedSchema
	err := env.dbs.Read.Get(&s, account.StmtJoinedAccountByWx, unionID)

	if err != nil {
		return account.JoinedSchema{}, err
	}

	return s, nil
}

type joinedSchemaAsyncResult struct {
	value account.JoinedSchema
	err   error
}

func (env Env) AsyncJoinedByFtcID(ftcID string) <-chan joinedSchemaAsyncResult {
	c := make(chan joinedSchemaAsyncResult)

	go func() {
		defer close(c)
		s, err := env.JoinedByFtcID(ftcID)

		c <- joinedSchemaAsyncResult{
			value: s,
			err:   err,
		}
	}()

	return c
}

func (env Env) AsyncJoinedByWx(unionID string) <-chan joinedSchemaAsyncResult {
	c := make(chan joinedSchemaAsyncResult)

	go func() {
		defer close(c)

		s, err := env.JoinedByWxID(unionID)

		c <- joinedSchemaAsyncResult{
			value: s,
			err:   err,
		}
	}()

	return c
}
