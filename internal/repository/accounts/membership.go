package accounts

import (
	"database/sql"
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type memberAsyncResult struct {
	value reader.Membership
	err   error
}

func (env Env) AsyncLoadMembership(compoundID string) <-chan memberAsyncResult {
	c := make(chan memberAsyncResult)

	go func() {
		m, err := env.RetrieveMember(compoundID)

		c <- memberAsyncResult{
			value: m,
			err:   err,
		}
	}()

	return c
}

func (env Env) countMemberSnapshot(ids ids.UserIDs) (int64, error) {
	var count int64
	err := env.DBs.Read.Get(
		&count,
		reader.StmtCountSnapshot,
		ids.BuildFindInSet(),
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (env Env) listMemberSnapshot(ids ids.UserIDs, p gorest.Pagination) ([]reader.MemberSnapshot, error) {
	var s = make([]reader.MemberSnapshot, 0)
	err := env.DBs.Read.Select(
		&s,
		reader.StmtListSnapshots,
		ids.BuildFindInSet(),
		p.Limit,
		p.Offset(),
	)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (env Env) ListSnapshot(ids ids.UserIDs, p gorest.Pagination) (reader.SnapshotList, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan reader.SnapshotList)

	go func() {
		defer close(countCh)
		n, err := env.countMemberSnapshot(ids)
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		s, err := env.listMemberSnapshot(ids, p)
		if err != nil {
			sugar.Error(err)
		}
		listCh <- reader.SnapshotList{
			Total:      0,
			Pagination: gorest.Pagination{},
			Data:       s,
			Err:        err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return reader.SnapshotList{}, listResult.Err
	}

	return reader.SnapshotList{
		Total:      count,
		Pagination: p,
		Data:       listResult.Data,
	}, nil
}

// CreateMembership manually.
func (env Env) CreateMembership(ba account.BaseAccount, params input.MemberParams) (reader.Membership, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

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

func (env Env) UpdateMembership(compoundID string, params input.MemberParams) (reader.MembershipVersioned, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

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

	return updated.Version(reader.Archiver{
		Name:   reader.ArchiveName(params.CreatedBy),
		Action: reader.ArchiveActionUpdate,
	}).WithPriorVersion(current), nil
}

func (env Env) DeleteMembership(compoundID string) (reader.Membership, error) {
	defer env.Logger.Sync()
	sugar := env.Logger.Sugar()

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
		return m, nil
	}

	if !m.IsOneTime() {
		return reader.Membership{}, errors.New("only one-time purchase membership could be deleted directly")
	}

	err = tx.DeleteMember(m.UserIDs)

	if err != nil {
		return reader.Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.Membership{}, err
	}

	return m, nil
}
