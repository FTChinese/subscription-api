package readers

import (
	"database/sql"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// RetrieveMember loads reader.Membership of the specified id.
// compoundID - Might be ftc uuid or wechat union id.
func (env Env) RetrieveMember(compoundID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtSelectMember,
		compoundID)

	if err != nil && err != sql.ErrNoRows {
		return reader.Membership{}, err
	}

	return m.Sync(), nil
}

// RetrieveAppleMember selects membership by apple original transaction id.
// // NOTE: sql.ErrNoRows are ignored. The returned
//// Membership might be a zero value.
func (env Env) RetrieveAppleMember(txID string) (reader.Membership, error) {
	var m reader.Membership

	err := env.DBs.Read.Get(
		&m,
		reader.StmtAppleMember,
		txID)

	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	return m.Sync(), nil
}

// ArchiveMember saves a member's snapshot at a specific moment.
// Deprecated.
func (env Env) ArchiveMember(snapshot reader.MemberSnapshot) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtSaveSnapshot,
		snapshot)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) VersionMembership(v reader.MembershipVersioned) error {
	_, err := env.DBs.Write.NamedExec(
		reader.StmtVersionMembership,
		v)

	if err != nil {
		return err
	}

	return nil
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
