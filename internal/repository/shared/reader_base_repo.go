package shared

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ReaderBaseRepo contains shared functionalities of a reader.
// It contains methods to retrieve user's
// basic account data using various id fields.
// It also contains methods to retrieve membership
// using various ids.
type ReaderBaseRepo struct {
	DBs db.ReadWriteMyDBs
}

func NewReaderBaseRepo(dbs db.ReadWriteMyDBs) ReaderBaseRepo {
	return ReaderBaseRepo{
		DBs: dbs,
	}
}

func (env ReaderBaseRepo) BaseAccountByUUID(id string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByUUID, id)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

func (env ReaderBaseRepo) BaseAccountByStripeID(cusID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountOfStripe, cusID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// BaseAccountByWxID retrieves BaseAccount for a wechat user.
func (env ReaderBaseRepo) BaseAccountByWxID(unionID string) (account.BaseAccount, error) {
	var a account.BaseAccount
	err := env.DBs.Read.Get(&a, account.StmtBaseAccountByWx, unionID)
	if err != nil {
		return account.BaseAccount{}, err
	}

	return a, nil
}

// FindBaseAccount retrieve account by ftc id if exists, then fallback to union id.
func (env ReaderBaseRepo) FindBaseAccount(ids ids.UserIDs) (account.BaseAccount, error) {
	if ids.FtcID.Valid {
		return env.BaseAccountByUUID(ids.FtcID.String)
	}

	if ids.UnionID.Valid {
		return env.BaseAccountByWxID(ids.UnionID.String)
	}

	return account.BaseAccount{}, errors.New("either ftc id nor wechat id should be specified")
}

func (env ReaderBaseRepo) SearchUserByFtcOrWxID(id string) (account.BaseAccount, error) {
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
func (env ReaderBaseRepo) RetrieveMember(compoundID string) (reader.Membership, error) {
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
func (env ReaderBaseRepo) RetrieveAppleMember(txID string) (reader.Membership, error) {
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
func (env ReaderBaseRepo) ArchiveMember(snapshot reader.MemberSnapshot) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtSaveSnapshot,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

func (env ReaderBaseRepo) VersionMembership(v reader.MembershipVersioned) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtVersionMembership,
		v)

	if err != nil {
		return err
	}

	return nil
}
