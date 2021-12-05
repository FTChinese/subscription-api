package shared

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env ReaderCommon) JoinedByFtcID(ftcID string) (account.JoinedSchema, error) {
	var s account.JoinedSchema
	err := env.DBs.Read.Get(&s, account.StmtJoinedAccountByFtc, ftcID)
	if err != nil {
		return account.JoinedSchema{}, err
	}

	return s, nil
}

func (env ReaderCommon) JoinedByWxID(unionID string) (account.JoinedSchema, error) {
	var s account.JoinedSchema
	err := env.DBs.Read.Get(&s, account.StmtJoinedAccountByWx, unionID)

	if err != nil {
		return account.JoinedSchema{}, err
	}

	return s, nil
}

type joinedSchemaAsyncResult struct {
	value account.JoinedSchema
	err   error
}

func (env ReaderCommon) AsyncJoinedByFtcID(ftcID string) <-chan joinedSchemaAsyncResult {
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

func (env ReaderCommon) AsyncJoinedByWx(unionID string) <-chan joinedSchemaAsyncResult {
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

// AccountByFtcID retrieves reader account by ftc id.
// For compatibility, you should check where this user exists
// in profile table after account retrieved.
func (env ReaderCommon) AccountByFtcID(ftcID string) (reader.Account, error) {
	aChan, mChan := env.AsyncJoinedByFtcID(ftcID), env.AsyncLoadMembership(ftcID)

	aResult, mResult := <-aChan, <-mChan

	if aResult.err != nil {
		return reader.Account{}, aResult.err
	}

	if mResult.err != nil {
		return reader.Account{}, mResult.err
	}

	return reader.NewAccount(
		aResult.value,
		mResult.value,
		enum.LoginMethodEmail), nil
}

func (env ReaderCommon) AccountByWxID(unionID string) (reader.Account, error) {
	aChan, mChan := env.AsyncJoinedByWx(unionID), env.AsyncLoadMembership(unionID)

	aResult, mResult := <-aChan, <-mChan

	if aResult.err != nil {
		return reader.Account{}, aResult.err
	}

	if mResult.err != nil {
		return reader.Account{}, mResult.err
	}

	return reader.NewAccount(
		aResult.value,
		mResult.value,
		enum.LoginMethodWx), nil
}

type memberAsyncResult struct {
	value reader.Membership
	err   error
}

func (env ReaderCommon) AsyncLoadMembership(compoundID string) <-chan memberAsyncResult {
	c := make(chan memberAsyncResult)

	go func() {
		m, err := env.RetrieveMember(compoundID)

		c <- memberAsyncResult{
			value: m,
			err:   err,
		}
	}()

	return c
}
