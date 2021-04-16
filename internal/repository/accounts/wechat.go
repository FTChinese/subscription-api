package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// WxSignUp helps a Wechat-logged-in user to sign up on FTC.
// If user already purchased membership with Wechat account, the membership will be bound to this signup email.
// Returns the new account's UUID.
func (env Env) WxSignUp(unionID string, input pkg.EmailSignUpParams) (reader.LinkWxResult, error) {
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
		return reader.LinkWxResult{}, err
	}

	// Email already exists.
	if ok {
		return reader.LinkWxResult{}, render.NewVEAlreadyExists("email")
	}

	// Retrieve account by wx union id.
	wxAccount, err := env.AccountByWxID(unionID)
	if err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	merged, err := ftcAccount.Link(wxAccount)
	if err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	// Start persisting data.
	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	//
	if err = tx.CreateAccount(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	if err = tx.CreateProfile(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	if merged.Membership.IsZero() {
		if err := tx.Commit(); err != nil {
			sugar.Error(err)
			return reader.LinkWxResult{}, err
		}
		return reader.LinkWxResult{
			Account: merged,
		}, nil
	}

	sugar.Infof("Removing merged members...")
	if err := tx.DeleteMember(wxAccount.Membership.UserIDs); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	sugar.Infof("Inserting merged member...")
	if err := tx.CreateMember(merged.Membership); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	sugar.Infof("Wechat user %s has membership and is linked to new account %s", wxAccount.UnionID.String, ftcAccount.FtcID)

	if err := tx.Commit(); err != nil {
		sugar.Error()
		return reader.LinkWxResult{}, err
	}

	return reader.LinkWxResult{
		Account:           merged,
		FtcMemberSnapshot: reader.MemberSnapshot{},
		WxMemberSnapshot: wxAccount.Membership.Snapshot(reader.Archiver{
			Name:   reader.NameWechat,
			Action: reader.ActionLink,
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
func (env Env) LinkWechat(input pkg.LinkWxParams) (reader.LinkWxResult, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	// Retrieve accounts for ftc side and wx side respectively.
	ftcAcnt, err := env.AccountByFtcID(input.FtcID)
	if err != nil {
		return reader.LinkWxResult{}, err
	}

	wxAcnt, err := env.AccountByWxID(input.UnionID)
	if err != nil {
		return reader.LinkWxResult{}, err
	}

	merged, err := ftcAcnt.Link(wxAcnt)
	if err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	tx, err := env.beginAccountTx()
	if err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	if err := tx.AddUnionIDToFtc(merged.BaseAccount); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	// If both accounts do not have memberships, stop.
	if merged.Membership.IsZero() {
		sugar.Infof("Merged account have zero membership")
		if err := tx.Commit(); err != nil {
			return reader.LinkWxResult{}, err
		}

		return reader.LinkWxResult{
			Account: merged,
		}, nil
	}

	sugar.Infof("Removing merged members...")
	if err := tx.DeleteMember(merged.Membership.UserIDs); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	sugar.Infof("Inserting merged member...")
	if err := tx.CreateMember(merged.Membership); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.LinkWxResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return reader.LinkWxResult{}, err
	}

	return reader.LinkWxResult{
		Account: merged,
		FtcMemberSnapshot: ftcAcnt.Membership.Snapshot(reader.Archiver{
			Name:   reader.NameWechat,
			Action: reader.ActionLink,
		}),
		WxMemberSnapshot: wxAcnt.Membership.Snapshot(reader.Archiver{
			Name:   reader.NameWechat,
			Action: reader.ActionLink,
		}),
	}, nil
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
			Name:   reader.NameWechat,
			Action: reader.ActionUnlink,
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
