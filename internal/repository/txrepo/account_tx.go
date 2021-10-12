package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/account"
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

// RetrieveMobiles retrieves all rows matching a ftc id
// or a mobile number.
// Returns two row at maximum.
// If two rows returned, in indicates the ftc id and mobile
// does not belong the same account. You should not link the
// passed-in MobileUpdater.
// If zero row returned, you are safe to insert the MobileUpdater.
// If one row returned, you could only update it with
// the provided mobile if this row does not have mobile set
// and the ftc id matches.
func (tx AccountTx) RetrieveMobiles(u account.MobileUpdater) ([]account.MobileUpdater, error) {
	var mobiles = make([]account.MobileUpdater, 0)
	err := tx.Select(
		&mobiles,
		account.StmtLockProfileByIDOrMobile,
		u.FtcID,
		u.Mobile)

	if err != nil {
		return nil, err
	}

	return mobiles, nil
}

// UpsertMobile inserts a row into profile table, or update the mobile_phone
// column if user id already exists
func (tx AccountTx) UpsertMobile(params account.MobileUpdater) error {
	_, err := tx.NamedExec(
		account.StmtUpsertPhone,
		params)

	if err != nil {
		return err
	}

	return nil
}
