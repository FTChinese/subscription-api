package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// AccountByFtcID retrieves reader account by ftc id.
// For compatibility you should check where this user exists
// in profile table after account retrieved.
func (env Env) AccountByFtcID(ftcID string) (reader.Account, error) {
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

func (env Env) AccountByWxID(unionID string) (reader.Account, error) {
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
