package shared

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ReaderCommon contains shared functionalities of a reader.
// It contains methods to retrieve user's
// basic account data using various id fields.
// It also contains methods to retrieve membership
// using various ids.
type ReaderCommon struct {
	DBs db.ReadWriteMyDBs
}

func NewReaderCommon(dbs db.ReadWriteMyDBs) ReaderCommon {
	return ReaderCommon{
		DBs: dbs,
	}
}

func (env ReaderCommon) BaseAccountByUUID(id string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByUUID, id)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env ReaderCommon) BaseAccountByStripeID(cusID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountOfStripe, cusID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// BaseAccountByWxID retrieves BaseAccount for a wechat user.
func (env ReaderCommon) BaseAccountByWxID(unionID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByWx, unionID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// FindBaseAccount retrieve account by ftc id if exists, then fallback to union id.
func (env ReaderCommon) FindBaseAccount(ids ids.UserIDs) (account.BaseAccount, error) {
	if ids.FtcID.Valid {
		return env.BaseAccountByUUID(ids.FtcID.String)
	}

	if ids.UnionID.Valid {
		return env.BaseAccountByWxID(ids.UnionID.String)
	}

	return account.BaseAccount{}, errors.New("either ftc id nor wechat id should be specified")
}

func (env ReaderCommon) SearchUserByFtcOrWxID(id string) (account.BaseAccount, error) {
	ba, err := env.BaseAccountByUUID(id)
	if err == nil {
		return ba, nil
	}

	if err != sql.ErrNoRows {
		return account.BaseAccount{}, err
	}

	return env.BaseAccountByWxID(id)
}

// RetrieveMember loads reader.Membership of the specified id.
// compoundID - Might be ftc uuid or wechat union id.
func (env ReaderCommon) RetrieveMember(compoundID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtSelectMember,
		compoundID)

	if err != nil && err != sql.ErrNoRows {
		return reader.Membership{}, err
	}

	return m.Sync(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (env ReaderCommon) RetrieveAppleMember(txID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtAppleMember,
		txID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

// ArchiveMember saves a member's snapshot at a specific moment.
// Deprecated.
func (env ReaderCommon) ArchiveMember(snapshot reader.MemberSnapshot) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtSaveSnapshot,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

func (env ReaderCommon) VersionMembership(v reader.MembershipVersioned) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtVersionMembership,
		v)

	if err != nil {
		return err
	}

	return nil
}
