package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/jmoiron/sqlx"
)

type AccountTx struct {
	SharedTx
}

func NewAccountTx(tx *sqlx.Tx) AccountTx {
	return AccountTx{
		SharedTx: NewSharedTx(tx),
	}
}

// CreateAccount create a ftc account.
func (tx AccountTx) CreateAccount(a account.BaseAccount) error {

	_, err := tx.NamedExec(
		account.StmtCreateFtc,
		a)

	if err != nil {
		return err
	}

	return nil
}

// CreateProfile creates profile when creating ftc account.
func (tx AccountTx) CreateProfile(a account.BaseAccount) error {
	_, err := tx.NamedExec(
		account.StmtCreateProfile,
		a)

	if err != nil {
		return err
	}

	return nil
}

// AddUnionIDToFtc links two existing account by setting wechat union id to userinfo table.
func (tx AccountTx) AddUnionIDToFtc(a account.BaseAccount) error {

	_, err := tx.NamedExec(
		account.StmtLinkAccount,
		a)

	if err != nil {
		return err
	}

	return nil
}

func (tx AccountTx) RetrieveMobile(ftcID string) (ztsms.MobileUpdater, error) {
	var p ztsms.MobileUpdater
	err := tx.Get(&p, ztsms.StmtLockMobileByID, ftcID)

	if err != nil {
		return ztsms.MobileUpdater{}, err
	}

	return p, nil
}

func (tx AccountTx) SetMobile(params ztsms.MobileUpdater) error {
	_, err := tx.NamedExec(
		ztsms.StmtSetPhone,
		params)

	if err != nil {
		return err
	}

	return nil
}
