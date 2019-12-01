package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
)

func GenerateSnapshotID() string {
	return "snp_" + rand.String(12)
}

// MemberSnapshot saves a membership's status prior to
// placing an order.
// TODO: rename ID to avoid conflict.
type MemberSnapshot struct {
	SnapshotID string      `db:"snapshot_id"`
	Reason     SubsKind    `db:"reason"`
	CreatedUTC chrono.Time `db:"created_utc"`
	Membership
}

func NewMemberSnapshot(m Membership, reason SubsKind) MemberSnapshot {
	return MemberSnapshot{
		SnapshotID: GenerateSnapshotID(),
		Reason:     reason,
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
