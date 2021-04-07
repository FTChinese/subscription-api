package accounts

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
)

// CreateAccount creates a new new FTC account.
func (env Env) CreateAccount(a account.BaseAccount) error {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	err = tx.CreateAccount(a)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	err = tx.CreateProfile(a)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

func (env Env) BaseAccountByUUID(id string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.dbs.Read.Get(&a, account.StmtBaseAccountByUUID, id)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env Env) BaseAccountByWxID(unionID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.dbs.Read.Get(&a, account.StmtBaseAccountByWx, unionID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// FindBaseAccount retrieve account by ftc id if exists, then fallback to union id.
func (env Env) FindBaseAccount(ids pkg.UserIDs) (account.BaseAccount, error) {
	if ids.FtcID.Valid {
		return env.BaseAccountByUUID(ids.FtcID.String)
	}

	if ids.UnionID.Valid {
		return env.BaseAccountByWxID(ids.UnionID.String)
	}

	return account.BaseAccount{}, errors.New("neither ftc id nor wechat id should be specified")
}

func (env Env) BaseAccountByMobile(mobile string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.dbs.Read.Get(&a, account.StmtBaseAccountByMobile, mobile)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env Env) BaseAccountByStripeID(cusID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.dbs.Read.Get(&a, account.StmtBaseAccountOfStripe, cusID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}
