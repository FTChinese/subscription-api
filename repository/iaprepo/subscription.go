package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// UpsertSubscription saves an Subscription instance
// built from the latest transaction, or update it if exists.
func (env Env) UpsertSubscription(s apple.Subscription) error {
	_, err := env.db.NamedExec(apple.StmtUpsertSubs, s)

	if err != nil {
		return err
	}

	return nil
}

// UpdateMembership update subs.Membership if it exists.
// Return a subs.MemberSnapshot if this subscription is linked to ftc account; otherwise it is empty.
func (env Env) UpdateMembership(s apple.Subscription) (subs.MemberSnapshot, error) {
	tx, err := env.BeginTx()
	if err != nil {
		logger.Error(err)
		return subs.MemberSnapshot{}, err
	}

	// Retrieve membership by original transaction id.
	// This is the only information we know about a possible user.
	// If the membership is not found, we can assume this IAP is not linked to FTC account.
	currMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		logger.Error(err)
		_ = tx.Rollback()
		return subs.MemberSnapshot{}, err
	}

	// If the subscription is not linked to FTC account, return empty MemberSnapshot and not error.
	if currMember.IsZero() {
		_ = tx.Commit()
		return subs.MemberSnapshot{}, nil
	}

	newMember := s.BuildOn(currMember)

	if err := tx.UpdateMember(newMember); err != nil {
		_ = tx.Rollback()
		return subs.MemberSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return subs.MemberSnapshot{}, err
	}

	return currMember.Snapshot(enum.SnapshotReasonAppleLink), nil
}

func (env Env) LoadSubscription(originalID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := env.db.Get(&s, apple.StmtLoadSubs, originalID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}
