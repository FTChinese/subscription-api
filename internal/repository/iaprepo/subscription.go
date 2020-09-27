package iaprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"log"
)

// SaveSubs saves an apple.Subscription instance and
// optionally updated membership if it is linked to a ftc membership.
// This is used by verify receipt, refresh subscription, webhook, and polling.
// The returned membership is empty if the subscription is not linked to an FTC account.
func (env Env) SaveSubs(s apple.Subscription) (reader.MemberSnapshot, error) {
	err := env.upsertSubscription(s)
	if err != nil {
		return reader.MemberSnapshot{}, err
	}

	snapshot, err := env.updateMembership(s)
	if err != nil {
		return reader.MemberSnapshot{}, err
	}

	return snapshot, nil
}

// upsertSubscription saves an Subscription instance
// built from the latest transaction, or update it if exists.
func (env Env) upsertSubscription(s apple.Subscription) error {
	_, err := env.db.NamedExec(apple.StmtUpsertSubs, s)

	if err != nil {
		return err
	}

	return nil
}

// updateMembership update subs.Membership if it is linked to an apple subscription.
// Return a subs.MemberSnapshot if this subscription is linked to ftc account; otherwise it is empty.
func (env Env) updateMembership(s apple.Subscription) (reader.MemberSnapshot, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return reader.MemberSnapshot{}, err
	}

	// Retrieve membership by original transaction id.
	// This is the only information we know about a possible user.
	// If the membership is not found, we can assume this IAP is not linked to FTC account.
	currMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	// If the subscription is not linked to FTC account, return empty MemberSnapshot and not error.
	if currMember.IsZero() {
		sugar.Info("Empty membership linked to IAP. Ignore membership update.")
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, nil
	}

	newMember := s.BuildOn(currMember)

	if err := tx.UpdateMember(newMember); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return reader.MemberSnapshot{}, err
	}

	return currMember.Snapshot(enum.SnapshotReasonIapUpdate), nil
}

func (env Env) LoadSubs(originalID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := env.db.Get(&s, apple.StmtLoadSubs, originalID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}

func (env Env) countSubs() (int64, error) {
	var count int64
	err := env.db.Get(&count, apple.StmtCountSubs)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (env Env) listSubs(p gorest.Pagination) ([]apple.Subscription, error) {
	var s = make([]apple.Subscription, 0)
	err := env.db.Select(&s, apple.StmtListSubs, p.Limit, p.Offset())
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (env Env) ListSubs(p gorest.Pagination) (apple.SubsList, error) {
	countCh := make(chan int64)
	listCh := make(chan apple.SubsList)

	go func() {
		defer close(countCh)
		n, err := env.countSubs()
		// Ignore error
		if err != nil {
			log.Print(err)
		}
		countCh <- n
	}()

	go func() {
		defer close(listCh)
		s, err := env.listSubs(p)
		listCh <- apple.SubsList{
			Total:      0,
			Pagination: gorest.Pagination{},
			Data:       s,
			Err:        err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return apple.SubsList{}, listResult.Err
	}

	return apple.SubsList{
		Total:      count,
		Pagination: p,
		Data:       listResult.Data,
	}, nil
}
