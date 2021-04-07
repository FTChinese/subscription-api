package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

func (env Env) AccountByFtcID(ftcID string) (reader.Account, error) {
	aChan, mChan := env.AsyncJoinedByFtcID(ftcID), env.AsyncMembership(pkg.UserIDs{
		CompoundID: "",
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}.MustNormalize())

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
	aChan, mChan := env.AsyncJoinedByWx(unionID), env.AsyncMembership(pkg.UserIDs{
		CompoundID: "",
		FtcID:      null.String{},
		UnionID:    null.StringFrom(unionID),
	}.MustNormalize())

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
