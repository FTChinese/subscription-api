package reader

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/guregu/null"
)

type Account struct {
	account.BaseAccount
	LoginMethod enum.LoginMethod `json:"loginMethod"`
	Wechat      account.Wechat   `json:"wechat"`
	Membership  Membership       `json:"membership"`
}

func NewAccount(s account.JoinedSchema, m Membership, method enum.LoginMethod) Account {
	if s.VIP {
		m.Tier = enum.TierPremium
		m.VIP = true
	}

	return Account{
		BaseAccount: s.BaseAccount,
		LoginMethod: method,
		Wechat:      s.Wechat,
		Membership:  m,
	}
}

// IsLinked checks whether this account is linked to wechat account.
// Both FTC id and union id should exist.
func (a Account) IsLinked() bool {
	return a.FtcID != "" && a.UnionID.Valid
}

// IsEqual checks whether two accounts are the same one.
func (a Account) IsEqual(other Account) bool {
	return a.IsLinked() && other.IsLinked() && a.FtcID == other.FtcID
}

// HasMember checks whether this account has a membership, regardless of if it is valid or not.
func (a Account) HasMember() bool {
	return !a.Membership.IsZero()
}

// ValidateUnlink tests whether an account is allowed to
// be unlinked.
func (a Account) ValidateUnlink(anchor enum.AccountKind) *render.ValidationError {
	if !a.HasMember() {
		return nil
	}

	if anchor == enum.AccountKindNull {
		return &render.ValidationError{
			Message: "unlinking accounts must specify to which account membership should be reserved",
			Field:   "anchor",
			Code:    render.CodeMissingField,
		}
	}

	if a.Membership.IsFtcOnly() && anchor != enum.AccountKindFtc {
		return &render.ValidationError{
			Message: "valid stripe or iap subscription cannot transfer membership to wechat account upon unlinking",
			Field:   "anchor",
			Code:    render.CodeInvalid,
		}
	}

	if a.IsTest() && anchor != enum.AccountKindFtc {
		return &render.ValidationError{
			Message: "Test account cannot transfer membership to wechat account",
			Field:   "anchor",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}

// Link merges two accounts and their memberships if allowed.
func (a Account) Link(other Account) (Account, error) {
	// Noop.
	if a.IsEqual(other) {
		return Account{}, ErrAccountsAlreadyLinked
	}

	// If any of the two accounts is linked to a third one.
	if a.IsLinked() || other.IsLinked() {
		return Account{}, &render.ValidationError{
			Message: "one of the accounts or both of them are linked to a 3rd account",
			Field:   "account_link",
			Code:    render.CodeAlreadyExists,
		}
	}

	// Merge membership
	mm, err := a.Membership.Merge(other.Membership)
	if err != nil {
		return Account{}, err
	}

	// Start copying data. Use the FTC account target and
	// Wechat side as source since it involves less operation.
	// If a is FTC account, copy other's UnionID and Wechat fields.
	var merged Account
	if a.LoginMethod == enum.LoginMethodEmail {
		// For wechat new signup, this will other copy
		// password.
		merged = a
		merged.UnionID = other.UnionID
		merged.Wechat = other.Wechat
	} else {
		// Otherwise other is FTC account, copy a's
		merged = other
		merged.UnionID = a.UnionID
		merged.Wechat = a.Wechat
	}

	// Only set member ids for non-zero value.
	if !mm.IsZero() {
		mm.FtcID = null.StringFrom(merged.FtcID)
		mm.UnionID = merged.UnionID
		mm.CompoundID = mm.FtcID.String
	}

	merged.Membership = mm

	return merged, nil
}
