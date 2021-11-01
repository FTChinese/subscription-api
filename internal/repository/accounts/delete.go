package accounts

import "github.com/FTChinese/subscription-api/pkg/account"

func (env Env) DeleteAccount(d account.DeletedUser) error {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	err = tx.DeleteUserInfo(d.ID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return err
	}

	err = tx.DeleteProfile(d.ID)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return err
	}

	err = tx.SaveDeletedUser(d)
	if err != nil {
		_ = tx.Rollback()
		sugar.Error(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
