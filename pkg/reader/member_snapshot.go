package reader

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
)

type ArchiveName string

const (
	NameOrder  ArchiveName = "order"
	NameWechat ArchiveName = "wechat"
	NameApple  ArchiveName = "apple"
	NameStripe ArchiveName = "stripe"
	NameB2B    ArchiveName = "b2b"
)

type ArchiveAction string

const (
	ActionCreate     ArchiveAction = "create"  // Alipay, wechat, stripe
	ActionRenew      ArchiveAction = "renew"   // Alipay, wechat
	ActionUpgrade    ArchiveAction = "upgrade" // Alipay, wechat, stripe
	ActionAddOn      ArchiveAction = "transfer_addon"
	ActionVerify     ArchiveAction = "verify"  // Apple
	ActionPoll       ArchiveAction = "poll"    // Apple, alipay, wechat
	ActionLink       ArchiveAction = "link"    // Apple
	ActionUnlink     ArchiveAction = "unlink"  // Apple
	ActionRefresh    ArchiveAction = "refresh" // Stripe refresh.
	ActionCancel     ArchiveAction = "cancel"
	ActionReactivate ArchiveAction = "reactivate"
	ActionWebhook    ArchiveAction = "webhook" // Apple, Stripe webhook
)

type Archiver struct {
	Name   ArchiveName
	Action ArchiveAction
}

var (
	ArchiverAppleLink = Archiver{
		Name:   NameApple,
		Action: ActionLink,
	}
	ArchiverAppleUnlink = Archiver{
		Name:   NameApple,
		Action: ActionUnlink,
	}
)

func FtcArchiver(k enum.OrderKind) Archiver {
	switch k {
	case enum.OrderKindCreate:
		return Archiver{
			Name:   NameOrder,
			Action: ActionCreate,
		}

	case enum.OrderKindRenew:
		return Archiver{
			Name:   NameOrder,
			Action: ActionRenew,
		}

	case enum.OrderKindUpgrade:
		return Archiver{
			Name:   NameOrder,
			Action: ActionUpgrade,
		}

	case enum.OrderKindAddOn:
		return Archiver{
			Name:   NameOrder,
			Action: ActionAddOn,
		}
	}

	return Archiver{
		Name:   NameOrder,
		Action: "Unknown",
	}
}

func StripeArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   NameStripe,
		Action: a,
	}
}

func AppleArchiver(a ArchiveAction) Archiver {
	return Archiver{
		Name:   NameApple,
		Action: a,
	}
}

func (a Archiver) String() string {
	return fmt.Sprintf("%s.%s", a.Name, a.Action)
}

const StmtSnapshotMember = `
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

// MemberSnapshot saves a membership's status prior to
// placing an order.
type MemberSnapshot struct {
	SnapshotID string      `db:"snapshot_id"`
	CreatedBy  null.String `db:"created_by"`
	CreatedUTC chrono.Time `db:"created_utc"`
	OrderID    null.String `db:"order_id"` // Only exists when user is performing renewal or upgrading.
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
		SnapshotID: pkg.SnapshotID(),
		CreatedBy:  null.StringFrom(by.String()),
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
