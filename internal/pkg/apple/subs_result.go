package apple

import (
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type SubsResultParams struct {
	UserID        ids.UserIDs
	CurrentMember reader.Membership
}

type SubsResult struct {
	Subs             Subscription               `json:"subscription"`
	Member           reader.Membership          `json:"membership"`
	Versioned        reader.MembershipVersioned `json:"-"`
	CarryOverInvoice invoice.Invoice            `json:"-"`
}

func NewSubsResult(subs Subscription, params SubsResultParams) (SubsResult, error) {
	intent := reader.NewCheckoutIntentApple(params.CurrentMember)
	if intent.Error != nil {
		return SubsResult{}, intent.Error
	}

	// Webhook should never trigger this:
	// For new subscription we do no know user's current membership;
	// For existing membership, it cannot be purchased by means other than Apple.
	// This could only be possible when performing link.
	var inv invoice.Invoice
	if intent.Kind.IsSwitchToAutoRenew() {
		inv = params.CurrentMember.CarryOverInvoice().
			WithAppleTxID(subs.OriginalTransactionID)
	}

	m := NewMembership(MembershipParams{
		UserID: params.UserID,
		Subs:   subs,
		AddOn: params.CurrentMember.
			AddOn.
			Plus(addon.New(inv.Tier, inv.TotalDays())),
	})

	var versioned reader.MembershipVersioned
	if !params.CurrentMember.IsZero() {
		versioned = reader.NewMembershipVersioned(m).
			WithPriorVersion(params.CurrentMember).
			ArchivedBy(reader.NewArchiver().ByApple().ActionLink())
	}

	return SubsResult{
		Subs:             subs,
		Member:           m,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}, nil
}
