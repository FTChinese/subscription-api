package accounts

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
)

func (env Env) SaveSMSVerifier(v ztsms.Verifier) error {
	_, err := env.DBs.Write.NamedExec(ztsms.StmtSaveVerifier, v)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveSMSVerifier(params ztsms.VerifierParams) (ztsms.Verifier, error) {
	var v ztsms.Verifier
	err := env.DBs.Read.Get(&v, ztsms.StmtRetrieveVerifier, params.Mobile, params.Code)
	if err != nil {
		return ztsms.Verifier{}, err
	}

	return v, nil
}

func (env Env) SMSVerifierUsed(v ztsms.Verifier) error {
	_, err := env.DBs.Write.NamedExec(ztsms.StmtVerifierUsed, v)

	if err != nil {
		return err
	}

	return nil
}

// UpsertMobile inserts a new row in profile table or set
// mobile phone field if empty.
// Possibilities when you are trying to set the phone number:
// * The row with this user id does not exist at all.
//   In such case you should insert a row directly;
// * The row with this user id exists:
//   * If this row does not have mobile phone set, set it;
//   * If this row does have mobile phone set:
//     * If existing mobile matches the required one, stop;
//     * If existing mobile does not match the required one, it's conflict error.
func (env Env) UpsertMobile(params ztsms.MobileUpdater) error {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Retrieve a row matching the specified ftc id from profile table.
	// If the row does not exist, get a zero value.
	current, err := tx.RetrieveMobile(params.FtcID)
	if err != nil {
		if err != sql.ErrNoRows {
			sugar.Error(err)
			_ = tx.Rollback()
			return err
		}
		// Fallthrough for error no rows.
	}

	// If current.Mobile exists, the row must exist.
	// If mobile already exists.
	if current.Mobile.Valid {
		_ = tx.Rollback()
		// This ftc id already has this mobile set.
		if current.Mobile.String == params.Mobile.String {
			return errors.New("mobile already set")
		} else {
			return errors.New("already taken by another mobile")
		}
	}

	// Now the Mobile field is zero.
	// There are two cases here:
	// 1. The row does not exist, insert a new row into profile;
	// 2. The row exists but mobile is not set, update it.
	err = tx.SetMobile(params)
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

func (env Env) DeleteMobile(params ztsms.MobileUpdater) error {
	_, err := env.DBs.Write.Exec(ztsms.StmtUnsetMobile, params.FtcID)
	if err != nil {
		return err
	}

	return nil
}
