package txrepo

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
)

// UnlinkTx perform unlinking wechat account from ftc
// in a transaction.
type UnlinkTx struct {
	*sqlx.Tx
}

// UnlinkUser severs link between ftc account and
// wechat account by removing wx union id from user info.
func (tx UnlinkTx) UnlinkUser(a account.BaseAccount) error {
	_, err := tx.NamedExec(account.StmtUnlinkAccount, a)

	if err != nil {
		return err
	}

	return nil
}

// UnlinkMember an account's membership, if it exists.
// Returns ErrUnlinkAnchor
func (tx UnlinkTx) UnlinkMember(m reader.Membership, anchor enum.AccountKind) error {
	switch anchor {
	case enum.AccountKindFtc:
		err := tx.dropWxFromMember(m)
		if err != nil {
			return err
		}
		return nil

	case enum.AccountKindWx:
		err := tx.dropFtcFromMember(m)
		if err != nil {
			return err
		}
		return nil

	default:
		return errors.New("no idea which side should keep your membership after accounts unlinked")
	}
}

// dropWxFromMember strip union id from membership.
func (tx UnlinkTx) dropWxFromMember(m reader.Membership) error {
	_, err := tx.NamedExec(
		reader.StmtDropMemberUnionID,
		m)

	if err != nil {
		return err
	}

	return nil
}

// dropFtcFromMember strips ftc id from membership
func (tx UnlinkTx) dropFtcFromMember(m reader.Membership) error {
	_, err := tx.NamedExec(
		reader.StmtDropMemberFtcID,
		m)

	if err != nil {
		return err
	}

	return nil
}
