package readers

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ids"
)

func (env Env) BaseAccountByUUID(id string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByUUID, id)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env Env) BaseAccountByStripeID(cusID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountOfStripe, cusID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// BaseAccountByWxID retrieves BaseAccount for a wechat user.
func (env Env) BaseAccountByWxID(unionID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByWx, unionID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// FindBaseAccount retrieve account by ftc id if exists, then fallback to union id.
func (env Env) FindBaseAccount(ids ids.UserIDs) (account.BaseAccount, error) {
	if ids.FtcID.Valid {
		return env.BaseAccountByUUID(ids.FtcID.String)
	}

	if ids.UnionID.Valid {
		return env.BaseAccountByWxID(ids.UnionID.String)
	}

	return account.BaseAccount{}, errors.New("either ftc id nor wechat id should be specified")
}
