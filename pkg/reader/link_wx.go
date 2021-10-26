package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/guregu/null"
)

type WxEmailLinkResult struct {
	IsDuplicateLink   bool
	Account           Account // The account after linked
	FtcMemberSnapshot MemberSnapshot
	WxMemberSnapshot  MemberSnapshot
}

type WxEmailLinkBuilder struct {
	FTC    Account
	Wechat Account
}

// Link links an ftc account to wechat account.
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
func (b WxEmailLinkBuilder) Link() (Account, error) {
	// Noop.
	if b.FTC.IsEqual(b.Wechat) {
		return Account{}, ErrAccountsAlreadyLinked
	}

	// If any of the two accounts is linked to a third one.
	if b.FTC.IsLinked() || b.Wechat.IsLinked() {
		return Account{}, &render.ValidationError{
			Message: "one of the accounts or both of them are linked to a 3rd account",
			Field:   "account_link",
			Code:    render.CodeAlreadyExists,
		}
	}

	// Merge membership
	mm, err := b.FTC.Membership.Merge(b.Wechat.Membership)
	if err != nil {
		return Account{}, err
	}

	// Start copying data. Use the FTC account target and
	// Wechat side as source since it involves less operation.
	// If a is FTC account, copy other's UnionID and Wechat fields.
	merged := Account{
		BaseAccount: account.BaseAccount{
			FtcID:      b.FTC.FtcID,
			UnionID:    b.Wechat.UnionID,
			StripeID:   b.FTC.StripeID,
			Email:      b.FTC.Email,
			Mobile:     b.FTC.Mobile,
			UserName:   b.FTC.UserName,
			AvatarURL:  b.FTC.AvatarURL,
			IsVerified: b.FTC.IsVerified,
		},
		LoginMethod: 0,
		Wechat:      b.Wechat.Wechat,
		Membership:  Membership{},
	}

	// Only set member ids for non-zero value.
	if !mm.IsZero() {
		mm.CompoundID = merged.FtcID
		mm.FtcID = null.StringFrom(merged.FtcID)
		mm.UnionID = merged.UnionID
	}

	merged.Membership = mm

	return merged, nil
}

func (b WxEmailLinkBuilder) Build() (WxEmailLinkResult, error) {
	mergedAccount, err := b.Link()
	if err != nil {
		if errors.Is(err, ErrAccountsAlreadyLinked) {
			return WxEmailLinkResult{
				IsDuplicateLink:   true,
				Account:           b.FTC,
				FtcMemberSnapshot: MemberSnapshot{},
				WxMemberSnapshot:  MemberSnapshot{},
			}, nil
		}

		return WxEmailLinkResult{}, err
	}

	return WxEmailLinkResult{
		IsDuplicateLink: false,
		Account:         mergedAccount,
		FtcMemberSnapshot: b.FTC.Membership.Snapshot(Archiver{
			Name:   ArchiveNameWechat,
			Action: ActionActionLink,
		}),
		WxMemberSnapshot: b.Wechat.Membership.Snapshot(Archiver{
			Name:   ArchiveNameWechat,
			Action: ActionActionLink,
		}),
	}, nil
}
