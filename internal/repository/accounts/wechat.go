package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// WxSignUp helps a Wechat-logged-in user to sign up on FTC.
// If user already purchased membership with Wechat account, the membership will be bound to this signup email.
// Returns the new account's UUID.
func (env Env) WxSignUp(unionID string, input input.EmailSignUpParams) (reader.WxEmailLinkResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	// A new complete email account.
	// You should set LoginMethod to LoginMethodEmail
	// so that the Link step knows how to copy data.
	ftcAccount := reader.Account{
		BaseAccount: account.NewEmailBaseAccount(input),
		LoginMethod: enum.LoginMethodEmail,
		Wechat:      account.Wechat{},
		Membership:  reader.Membership{},
	}

	ok, err := env.EmailExists(input.Email)
	if err != nil {
		sugar.Error(err)
		return reader.WxEmailLinkResult{}, err
	}

	// Email already exists.
	if ok {
		return reader.WxEmailLinkResult{}, render.NewVEAlreadyExists("email")
	}

	// Retrieve account by wx union id.
	wxAccount, err := env.AccountByWxID(unionID)
	if err != nil {
		sugar.Error(err)
		return reader.WxEmailLinkResult{}, err
	}

	merged, err := ftcAccount.Link(wxAccount)
	if err != nil {
		sugar.Error(err)
		return reader.WxEmailLinkResult{}, err
	}

	// Start persisting data.
	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return reader.WxEmailLinkResult{}, err
	}

	//
	if err = tx.CreateAccount(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.WxEmailLinkResult{}, err
	}

	if err = tx.CreateProfile(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.WxEmailLinkResult{}, err
	}

	if merged.Membership.IsZero() {
		if err := tx.Commit(); err != nil {
			sugar.Error(err)
			return reader.WxEmailLinkResult{}, err
		}
		return reader.WxEmailLinkResult{
			Account: merged,
		}, nil
	}

	sugar.Infof("Removing merged members...")
	if err := tx.DeleteMember(wxAccount.Membership.UserIDs); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.WxEmailLinkResult{}, err
	}

	sugar.Infof("Inserting merged member...")
	if err := tx.CreateMember(merged.Membership); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.WxEmailLinkResult{}, err
	}

	sugar.Infof("Wechat user %s has membership and is linked to new account %s", wxAccount.UnionID.String, ftcAccount.FtcID)

	if err := tx.Commit(); err != nil {
		sugar.Error()
		return reader.WxEmailLinkResult{}, err
	}

	return reader.WxEmailLinkResult{
		Account:           merged,
		FtcMemberSnapshot: reader.MemberSnapshot{},
		WxMemberSnapshot: wxAccount.Membership.Snapshot(reader.Archiver{
			Name:   reader.ArchiveNameWechat,
			Action: reader.ActionActionLink,
		}),
	}, nil
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
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

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

	if !result.FtcMemberSnapshot.IsZero() {
		sugar.Infof("Removing ftc side members...")
		err := tx.DeleteMember(result.FtcMemberSnapshot.UserIDs)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return err
		}
	}

	if !result.WxMemberSnapshot.IsZero() {
		sugar.Infof("Removing wechat side members...")
		err := tx.DeleteMember(result.WxMemberSnapshot.UserIDs)
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

	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

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

	go func() {
		_ = env.ArchiveMember(acnt.Membership.Snapshot(reader.Archiver{
			Name:   reader.ArchiveNameWechat,
			Action: reader.ActionActionUnlink,
		}))
	}()

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
