package reader

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
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
	ArchiveActionCreate     ArchiveAction = "create"  // Alipay, wechat, stripe
	ArchiveActionRenew      ArchiveAction = "renew"   // Alipay, wechat
	ArchiveActionUpgrade    ArchiveAction = "upgrade" // Alipay, wechat, stripe
	ArchiveActionAddOn      ArchiveAction = "transfer_addon"
	ArchiveActionVerify     ArchiveAction = "verify"  // Apple
	ArchiveActionPoll       ArchiveAction = "poll"    // Apple, alipay, wechat
	ArchiveActionLink       ArchiveAction = "link"    // Apple
	ArchiveActionUnlink     ArchiveAction = "unlink"  // Apple
	ArchiveActionRefresh    ArchiveAction = "refresh" // Stripe refresh.
	ArchiveActionCancel     ArchiveAction = "cancel"
	ArchiveActionReactivate ArchiveAction = "reactivate"
	ArchiveActionWebhook    ArchiveAction = "webhook" // Apple, Stripe webhook
	ArchiveActionManual     ArchiveAction = "manual"
	ArchiveActionUpdate     ArchiveAction = "update"
	ArchiveActionDelete     ArchiveAction = "delete"
)

type Archiver struct {
	Name   ArchiveName
	Action ArchiveAction
}

func (a Archiver) String() string {
	return fmt.Sprintf("%s.%s", a.Name, a.Action)
}

func NewOrderArchiver(k enum.OrderKind) Archiver {
	switch k {
	case enum.OrderKindCreate:
		return NewFtcArchiver(ArchiveActionCreate)

	case enum.OrderKindRenew:
		return NewFtcArchiver(ArchiveActionRenew)

	case enum.OrderKindUpgrade:
		return NewFtcArchiver(ArchiveActionUpgrade)

	case enum.OrderKindAddOn:
		return NewFtcArchiver(ArchiveActionAddOn)
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
