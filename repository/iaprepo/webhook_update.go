package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// WebhookUpdate update membership after receiving data by webhook, if membership exists.
// Return a snapshot of membership before updating.
func (env Env) WebhookUpdate(s apple.Subscription) (subs.MemberSnapshot, error) {
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
