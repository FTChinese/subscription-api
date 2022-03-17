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
	Action        reader.ArchiveAction
}

type SubsResult struct {
	Subs             Subscription               `json:"subscription"`
	Member           reader.Membership          `json:"membership"`
	Versioned        reader.MembershipVersioned `json:"-"`
	CarryOverInvoice invoice.Invoice            `json:"-"`
}

func NewSubsResult(subs Subscription, params SubsResultParams) (SubsResult, error) {
	subsKind, err := params.CurrentMember.SubsKindByApple()
	if err != nil {
		return SubsResult{}, err
	}

	// Webhook should never trigger this:
	// For new subscription we do no know user's current membership;
	// For existing membership, it cannot be purchased by means other than Apple.
	// This could only be possible when performing link.
	var inv invoice.Invoice
	if subsKind == reader.SubsKindOneTimeToSub {
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
		m.Version(reader.NewAppleArchiver(params.Action)).
			WithPriorVersion(params.CurrentMember)
	}

	return SubsResult{
		Subs:             subs,
		Member:           m,
		Versioned:        versioned,
		CarryOverInvoice: inv,
	}, nil
}
