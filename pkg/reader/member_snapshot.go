package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/guregu/null"
)

func GenerateSnapshotID() string {
	return "snp_" + rand.String(12)
}

// MemberSnapshot saves a membership's status prior to
// placing an order.
type MemberSnapshot struct {
	SnapshotID string              `db:"snapshot_id"`
	Reason     enum.SnapshotReason `db:"reason"`
	CreatedUTC chrono.Time         `db:"created_utc"`
	OrderID    null.String         `db:"order_id"` // Only exists when user is performing renewal or upgrading.
	Membership
}

// GetSnapshotReason deduces why a membership is snapshot
// when an order is confirmed and membership updated.
func GetSnapshotReason(k enum.OrderKind) enum.SnapshotReason {
	switch k {
	case enum.OrderKindRenew:
		return enum.SnapshotReasonRenew

	case enum.OrderKindUpgrade:
		return enum.SnapshotReasonUpgrade

	default:
		return enum.SnapshotReasonNull
	}
}
