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
type MemberSnapshot struct {
	ID         string      `db:"snapshot_id"`
	CreatedUTC chrono.Time `db:"created_utc"`
	Membership
}

func NewMemberSnapshot(m Membership) MemberSnapshot {
	return MemberSnapshot{
		ID:         GenerateSnapshotID(),
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
