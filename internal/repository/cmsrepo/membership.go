package cmsrepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// CreateMembership manually.
func (env Env) CreateMembership(ba account.BaseAccount, params input.MemberParams) (reader.Membership, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginMemberTx()
	if err != nil {
		sugar.Error(err)
		return reader.Membership{}, err
	}

	mmb, err := tx.RetrieveMember(ba.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	if !mmb.IsZero() {
		_ = tx.Rollback()
		return reader.Membership{}, &render.ValidationError{
			Message: "Membership already exists",
			Field:   "membership",
			Code:    render.CodeAlreadyExists,
		}
	}

	mmb = reader.NewMembership(ba, params)

	err = tx.CreateMember(mmb)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.Membership{}, err
	}

	return mmb, nil
}

func (env Env) UpdateMembership(compoundID string, params input.MemberParams, by string) (reader.MembershipVersioned, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginMemberTx()
	if err != nil {
		sugar.Error(err)
		return reader.MembershipVersioned{}, err
	}

	current, err := tx.RetrieveMember(compoundID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MembershipVersioned{}, err
	}

	if current.IsZero() {
		_ = tx.Rollback()
		return reader.MembershipVersioned{}, sql.ErrNoRows
	}

	if !current.IsOneTime() {
		_ = tx.Rollback()
		return reader.MembershipVersioned{}, &render.ValidationError{
			Message: "Only membership created via alipay or wxpay can be modified directly",
			Field:   "payMethod",
			Code:    render.CodeAlreadyExists,
		}
	}

	updated := current.Update(params)

	err = tx.UpdateMember(updated)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MembershipVersioned{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.MembershipVersioned{}, err
	}

	v := reader.NewMembershipVersioned(updated).
		WithPriorVersion(current).
		ArchivedBy(reader.NewArchiver().By(by).ActionUpdate())
	return v, nil
}

func (env Env) DeleteMembership(compoundID string) (reader.Membership, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginMemberTx()
	if err != nil {
		sugar.Error(err)
		return reader.Membership{}, err
	}

	m, err := tx.RetrieveMember(compoundID)
	if err != nil {
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	if m.IsZero() {
		_ = tx.Rollback()
		return m, nil
	}

	if !m.IsOneTime() {
		_ = tx.Rollback()
		return reader.Membership{}, errors.New("only one-time purchase membership could be deleted directly")
	}

	err = tx.DeleteMember(m.UserIDs)

	if err != nil {
		_ = tx.Rollback()
		return reader.Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.Membership{}, err
	}

	return m, nil
}
