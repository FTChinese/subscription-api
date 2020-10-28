package reader

import (
	"database/sql/driver"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
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
	ActionCreate  ArchiveAction = "create"
	ActionRenew   ArchiveAction = "renew"
	ActionUpgrade ArchiveAction = "upgrade"
	ActionVerify  ArchiveAction = "verify"
	ActionPoll    ArchiveAction = "poll"
	ActionLink    ArchiveAction = "link"
	ActionUnlink  ArchiveAction = "unlink"
)

type Archiver struct {
	Name   ArchiveName
	Action ArchiveAction
}

var (
	ArchiverAppleVerify = Archiver{
		Name:   NameApple,
		Action: ActionVerify,
	}
	ArchiverApplePoll = Archiver{
		Name:   NameApple,
		Action: ActionPoll,
	}
	ArchiverAppleLink = Archiver{
		Name:   NameApple,
		Action: ActionLink,
	}
	ArchiverAppleUnlink = Archiver{
		Name:   NameApple,
		Action: ActionUnlink,
	}
	ArchiverStripeCreate = Archiver{
		Name:   NameStripe,
		Action: ActionCreate,
	}
	ArchiverStripeUpgrade = Archiver{
		Name:   NameStripe,
		Action: ActionUpgrade,
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
	}

	return Archiver{
		Name:   NameOrder,
		Action: "Unknown",
	}
}

func (a Archiver) String() string {
	return fmt.Sprintf("%a.%a", a.Name, a.Action)
}

func (a Archiver) Value() (driver.Value, error) {
	if a.Name == "" && a.Action == "" {
		return nil, nil
	}

	return a.String(), nil
}

func GenerateSnapshotID() string {
	return "snp_" + rand.String(12)
}

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
