package reader

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

type ArchiveName string

const (
	ArchiveNameOrder  ArchiveName = "order"  // For ftc order
	ArchiveNameWechat ArchiveName = "wechat" // For linking wechat
	ArchiveNameApple  ArchiveName = "apple"  // For apple link
	ArchiveNameStripe ArchiveName = "stripe"
	ArchiveNameB2B    ArchiveName = "b2b"
)

type ArchiveAction string

const (
	ActionActionCreate     ArchiveAction = "create"  // Alipay, wechat, stripe
	ActionActionRenew      ArchiveAction = "renew"   // Alipay, wechat
	ActionActionUpgrade    ArchiveAction = "upgrade" // Alipay, wechat, stripe
	ActionActionAddOn      ArchiveAction = "transfer_addon"
	ActionActionVerify     ArchiveAction = "verify"  // Apple
	ActionPoll             ArchiveAction = "poll"    // Apple, alipay, wechat
	ActionActionLink       ArchiveAction = "link"    // Apple
	ActionActionUnlink     ArchiveAction = "unlink"  // Apple
	ActionActionRefresh    ArchiveAction = "refresh" // Stripe refresh.
	ActionActionCancel     ArchiveAction = "cancel"
	ActionActionReactivate ArchiveAction = "reactivate"
	ActionActionWebhook    ArchiveAction = "webhook" // Apple, Stripe webhook
)

type Archiver struct {
	Name   ArchiveName
	Action ArchiveAction
}

func (a Archiver) String() string {
	return fmt.Sprintf("%s.%s", a.Name, a.Action)
}

var (
	ArchiverAppleLink = Archiver{
		Name:   ArchiveNameApple,
		Action: ActionActionLink,
	}
	ArchiverAppleUnlink = Archiver{
		Name:   ArchiveNameApple,
		Action: ActionActionUnlink,
	}
)

func NewOrderArchiver(k enum.OrderKind) Archiver {
	switch k {
	case enum.OrderKindCreate:
		return NewFtcArchiver(ActionActionCreate)

	case enum.OrderKindRenew:
		return NewFtcArchiver(ActionActionRenew)

	case enum.OrderKindUpgrade:
		return NewFtcArchiver(ActionActionUpgrade)

	case enum.OrderKindAddOn:
		return NewFtcArchiver(ActionActionAddOn)
	}

	return Archiver{
		Name:   ArchiveNameOrder,
		Action: "Unknown",
	}
}

func NewFtcArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   ArchiveNameOrder,
		Action: a,
	}
}

func NewWechatArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   ArchiveNameWechat,
		Action: a,
	}
}

func NewStripeArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   ArchiveNameStripe,
		Action: a,
	}
}

func NewAppleArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   ArchiveNameApple,
		Action: a,
	}
}

func NewB2BArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   ArchiveNameB2B,
		Action: a,
	}
}

const StmtSaveSnapshot = `
INSERT INTO premium.member_snapshot
SET id = :snapshot_id,
	created_by = :created_by,
	created_utc = UTC_TIMESTAMP(),
	order_id = :order_id,
	compound_id = :compound_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
	tier = :tier,
	cycle = :cycle,
` + mUpsertSharedCols

// StmtListSnapshots retrieves all membership change history.
// User might have ftc id, or union id, or both. We should
// retrieve all of them using the FIND_IN_SET function.
const StmtListSnapshots = `
SELECT id AS snapshot_id,
	created_by,
	created_utc,
	order_id,
	compound_id,
	ftc_user_id AS ftc_id,
	wx_union_id AS union_id,
	tier,
	cycle,
` + mRetrievalSharedCols + `
FROM premium.member_snapshot
WHERE FIND_IN_SET(compound_id, ?) > 0
ORDER BY created_utc DESC
LIMIT ? OFFSET ?`

const StmtCountSnapshot = `
SELECT COUNT(*) AS row_count
FROM premium.member_snapshot
WHERE FIND_IN_SET(compound_id, ?) > 0`

// MemberSnapshot saves a membership's status prior to
// placing an order.
// Deprecated. Will be saved in a single column as JSON.
type MemberSnapshot struct {
	SnapshotID string      `json:"id" db:"snapshot_id"`
	CreatedBy  null.String `json:"createdBy" db:"created_by"`
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	OrderID    null.String `json:"orderId" db:"order_id"` // Only exists when user is performing renewal or upgrading.
	Membership
}

func (s MemberSnapshot) WithOrder(id string) MemberSnapshot {
	s.OrderID = null.StringFrom(id)
	return s
}

func (s MemberSnapshot) WithArchiver(by Archiver) MemberSnapshot {
	s.CreatedBy = null.StringFrom(by.String())
	return s
}

// Snapshot takes a snapshot of membership, usually before modifying it.
func (m Membership) Snapshot(by Archiver) MemberSnapshot {
	if m.IsZero() {
		return MemberSnapshot{}
	}

	return MemberSnapshot{
		SnapshotID: ids.SnapshotID(),
		CreatedBy:  null.StringFrom(by.String()),
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}

type SnapshotList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []MemberSnapshot `json:"data"`
	Err  error            `json:"-"`
}
