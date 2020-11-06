package readerrepo

import (
	"database/sql"
	"errors"
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

func (env Env) FindAccount(ids reader.MemberID) (reader.FtcAccount, error) {
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
