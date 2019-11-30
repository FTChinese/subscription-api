package paywall

import (
	"github.com/FTChinese/go-rest/chrono"
	"gitlab.com/ftchinese/subscription-api/models/rand"
)

func GenerateSnapshotID() string {
	return "snp_" + rand.String(12)
}

// MemberSnapshot saves a membership's status prior to
// placing an order.
// TODO: rename ID to avoid conflict.
type MemberSnapshot struct {
	ID         string      `db:"snapshot_id"`
	Reason     SubsKind    `db:"reason"`
	CreatedUTC chrono.Time `db:"created_utc"`
	Membership
}

func NewMemberSnapshot(m Membership, reason SubsKind) MemberSnapshot {
	return MemberSnapshot{
		ID:         GenerateSnapshotID(),
		Reason:     reason,
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
