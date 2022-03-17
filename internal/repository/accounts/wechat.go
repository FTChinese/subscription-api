package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// WxSignUp helps a Wechat-logged-in user to sign up on FTC.
// If user already purchased membership with Wechat account, the membership will be bound to this signup email.
// Returns the new account's UUID.
func (env Env) WxSignUp(merged reader.Account) error {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	// Start persisting data.
	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	//
	if err = tx.CreateAccount(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err = tx.CreateProfile(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if merged.Membership.IsZero() {
		if err := tx.Commit(); err != nil {
			sugar.Error(err)
			return err
		}
		return nil
	}

	sugar.Infof("Removing merged members...")
	err = tx.DeleteMember(ids.UserIDs{
		CompoundID: "",
		FtcID:      null.String{},
		UnionID:    merged.UnionID,
	}.MustNormalize())

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	sugar.Infof("Inserting merged member...")
	if err := tx.CreateMember(merged.Membership); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	sugar.Infof("Wechat user %s has membership and is linked to new account %s", merged.UnionID.String, merged.FtcID)

	if err := tx.Commit(); err != nil {
		sugar.Error()
		return err
	}

	return nil
}

// LinkWechat links an ftc account to wechat account.
//
// Permissible merging matrix:
// FTC \ Wechat | no member | not expired | expired |
// no member    | Y         | Y            | Y      |
// no expired   | Y         | N            | Y      |
// expired      | Y         | Y            | Y      |
//
// To simplify calculation, we treat non-subscribed users as having a membership, which is the zero value of Membership.
// There's a special case caused by legacy behavior:
// Somehow (some might manually touched DB) the same reader's
// membership is linked while the accounts are not. We need to
// allow linking for such accounts.
func (env Env) LinkWechat(result reader.WxEmailLinkResult) error {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	if err := tx.AddUnionIDToFtc(result.Account.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	// If both accounts do not have memberships, stop.
	if result.Account.Membership.IsZero() {
		sugar.Infof("Merged account have zero membership")
		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}

	if !result.FtcVersioned.AnteChange.IsZero() {
		sugar.Infof("Removing ftc side members...")
		err := tx.DeleteMember(result.FtcVersioned.AnteChange.UserIDs)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return err
		}
	}

	if !result.WxVersioned.AnteChange.IsZero() {
		sugar.Infof("Removing wechat side members...")
		err := tx.DeleteMember(result.WxVersioned.AnteChange.UserIDs)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return err
		}
	}

	sugar.Infof("Inserting merged member...")
	if err := tx.CreateMember(result.Account.Membership); err != nil {
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

// UnlinkWx reverts linking ftc account to wechat account.
// First unlink membership if exists, then unlink account.
func (env Env) UnlinkWx(acnt reader.Account, anchor enum.AccountKind) error {

	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	ltx, err := env.beginUnlinkTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	err = ltx.UnlinkUser(acnt.BaseAccount)
	if err != nil {
		sugar.Error(err)
		_ = ltx.Rollback()
		return err
	}

	if acnt.Membership.IsZero() {
		if err := ltx.Commit(); err != nil {
			sugar.Error(err)
			return err
		}

		return nil
	}

	if err := ltx.UnlinkMember(acnt.Membership, anchor); err != nil {
		sugar.Error(err)
		_ = ltx.Rollback()
		return err
	}

	if err := ltx.Commit(); err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}
