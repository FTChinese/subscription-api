package cmsrepo

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// UpsertMembership manually create a new membership if not exists,
// or update an existing one.
func (env Env) UpsertMembership(m reader.Membership, by string) (reader.MembershipVersioned, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginMemberTx()
	if err != nil {
		sugar.Error(err)
		return reader.MembershipVersioned{}, err
	}

	current, err := tx.RetrieveMember(m.GetCompoundID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MembershipVersioned{}, err
	}

	if current.IsZero() {
		err = tx.CreateMember(m)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return reader.MembershipVersioned{}, err
		}
	} else {
		err = tx.UpdateMember(m)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return reader.MembershipVersioned{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return reader.MembershipVersioned{}, err
	}

	v := reader.NewMembershipVersioned(m).
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
