package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/account"
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
// * Both ftc id and mobile does not exist in table: insert directly.
// * Mobile exists:
//   - It is linked to another ftc id. This user is not allowed to use this mobile
//   - It is already linked to this ftc id. Do nothing.
// * Ftc ID exists:
//   - Its mobile_phone column is empty. Update it with this mobile.
//   - Its mobile_phone column is not empty and does not match this mobile number, indicating this ftc id already has a phone set.
//   - Its mobile_phone column is not empty and matches this one. Already linked and do nothing.
//
// In general, to set the mobile to this ftc id, we must make sure
// the mobile never appears in table.
func (env Env) UpsertMobile(params account.MobileUpdater) error {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	mobileRows, err := tx.RetrieveMobiles(params)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = account.IsMobileSettable(mobileRows, params)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// The table does not have this mobile number.
	// A row for this ftc id either does not exist,
	// or existed but mobile_phone column is empty.
	err = tx.UpsertMobile(params)
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

func (env Env) DeleteMobile(params account.MobileUpdater) error {
	_, err := env.DBs.Write.Exec(account.StmtUnsetMobile, params.FtcID)
	if err != nil {
		return err
	}

	return nil
}
