package iaprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SaveSubs saves an apple.Subscription instance and
// optionally update membership if it is linked to a ftc membership.
// This is used by verify receipt, refresh subscription, webhook, and polling.
// The returned membership is empty if the subscription is not linked to an FTC account.
func (env Env) SaveSubs(s apple.Subscription) (apple.SubsResult, error) {
	err := env.upsertSubscription(s)
	if err != nil {
		return apple.SubsResult{}, err
	}

	return env.updateMembership(s)
}

// upsertSubscription saves an Subscription instance
// built from the latest transaction, or update it if exists.
// Note the ftc_user_id field won't be touched here.
func (env Env) upsertSubscription(s apple.Subscription) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtUpsertSubs, s)

	if err != nil {
		return err
	}

	return nil
}

// updateMembership update subs.Membership if it is linked to an apple subscription,
// and returns a snapshot of membership if it is actually touched.
func (env Env) updateMembership(s apple.Subscription) (apple.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return apple.SubsResult{}, err
	}

	// Retrieve membership by original transaction id.
	// This is the only information we know about a possible user.
	// If the membership is not found, we can assume this IAP is not linked to FTC account.
	currMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.SubsResult{}, err
	}

	sugar.Infof("Membership linked to %s: %v", s.OriginalTransactionID, currMember)
	// If the subscription is not linked to FTC account, return empty MemberSnapshot and not error.
	// We need to take into account a situation where payment method is not apple but apple subscription id is not empty.
	if !s.ShouldUpdate(currMember) {
		sugar.Infof("Membership liked to original transaction id %s is either empty or non-iap, or expiration date not changed", s.OriginalTransactionID)
		_ = tx.Rollback()
		return apple.SubsResult{}, nil
	}

	// IAP membership exists. Update it.
	// Since we are polling IAP everyday, chances of membership
	// staying the same are every high.
	// We should compare the old and new memberships expiration time.
	sugar.Infof("Building membership based on %s", s.OriginalTransactionID)

	result, err := apple.NewSubsResult(s, apple.SubsResultParams{
		UserID:        currMember.UserIDs,
		CurrentMember: currMember,
		Action:        reader.ActionVerify,
	})
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.SubsResult{}, err
	}

	sugar.Infof("Membership %s expiration date updated from %s to %s", result.Member.CompoundID, currMember.ExpireDate, result.Member.ExpireDate)
	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.SubsResult{}, err
	}

	if !result.CarryOverInvoice.IsZero() {
		err := tx.SaveInvoice(result.CarryOverInvoice)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.SubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return apple.SubsResult{}, err
	}

	return result, nil
}

// LoadSubs retrieves a single row of iap subscription.
func (env Env) LoadSubs(originalID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := env.dbs.Read.Get(&s, apple.StmtLoadSubs, originalID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}

// countSubs get the number of rows of a user's subscription.
func (env Env) countSubs(ftcID string) (int64, error) {
	var count int64
	err := env.dbs.Read.Get(&count, apple.StmtCountSubs, ftcID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// listSubs get a list of a user's iap subscription.
func (env Env) listSubs(ftcID string, p gorest.Pagination) ([]apple.Subscription, error) {
	var s = make([]apple.Subscription, 0)
	err := env.dbs.Read.Select(&s, apple.StmtListSubs, ftcID, p.Limit, p.Offset())
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (env Env) ListSubs(ftcID string, p gorest.Pagination) (apple.SubsList, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan apple.SubsList)

	go func() {
		defer close(countCh)
		n, err := env.countSubs(ftcID)
		// Ignore error
		if err != nil {
			sugar.Error(err)
		}
		countCh <- n
	}()

	go func() {
		defer close(listCh)
		s, err := env.listSubs(ftcID, p)
		if err != nil {
			sugar.Error(err)
		}
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
