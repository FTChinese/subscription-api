package readerrepo

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) AccountByFtcID(id string) (reader.FtcAccount, error) {
	var u reader.FtcAccount
	err := env.db.Get(
		&u,
		reader.StmtAccountByFtcID,
		id,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (env Env) AccountByWxID(unionID string) (reader.FtcAccount, error) {
	var a reader.FtcAccount
	err := env.db.Get(&a, reader.StmtAccountByWx, unionID)
	if err != nil {
		return reader.FtcAccount{}, err
	}

	return a, nil
}

// FindAccount retrieve account by ftc id if exists, then fallback to union id.
func (env Env) FindAccount(ids pkg.MemberID) (reader.FtcAccount, error) {
	if ids.FtcID.Valid {
		return env.AccountByFtcID(ids.FtcID.String)
	}

	if ids.UnionID.Valid {
		return env.AccountByWxID(ids.UnionID.String)
	}

	return reader.FtcAccount{}, errors.New("niether ftc id nor wechat id should be specified")
}

func (env Env) FtcAccountByStripeID(cusID string) (reader.FtcAccount, error) {
	var u reader.FtcAccount
	err := env.db.Get(&u,
		reader.StmtAccountByStripeID,
		cusID)

	if err != nil {
		return u, err
	}

	return u, nil
}
