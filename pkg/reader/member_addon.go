package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"time"
)

// AddOnClaimed contains the data that should be
// used to perform add-on transfer.
type AddOnClaimed struct {
	Invoices   []invoice.Invoice // The invoices used to extend membership's expiration date. They will not be used next time.
	Membership Membership        // Update membership.
	Versioned  MembershipVersioned
}

// AddOnInvoiceCreated is the result of manually adding an invoice.
type AddOnInvoiceCreated struct {
	Invoice    invoice.Invoice     `json:"invoice"`
	Membership Membership          `json:"membership"`
	Versioned  MembershipVersioned `json:"versioned"`
}

func (m Membership) PlusAddOn(addOn addon.AddOn) Membership {
	m.AddOn = m.AddOn.Plus(addOn)
	return m
}

// ClearIAPWithAddOn generates an expired membership after user want to unlink
// IAP since the existence of addon prevents a simple deletion.
func (m Membership) ClearIAPWithAddOn() Membership {
	m.LegacyExpire = null.IntFrom(0)
	m.ExpireDate = chrono.Date{}
	m.PaymentMethod = enum.PayMethodAli
	m.FtcPlanID = null.String{}
	m.StripeSubsID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = false
	m.Status = enum.SubsStatusNull
	m.AppleSubsID = null.String{}
	m.B2BLicenceID = null.String{}

	return m
}

func (m Membership) CarriedOverAddOn() addon.AddOn {
	return addon.New(m.Tier, m.RemainingDays())
}

// NextRoundAddOn calculates the total addOn for such a
// situation:
// Current membership is converted to carry-over invoice;
// however, it already has addOn present. For the next
// round of membership to be generated, we have to add
// the carry-over's total days to current addon,
// which is equivalent to add current addon with remaining days.
// We should not use RemainingDays directly since current membership might
// not eligible for carry-over.
func (m Membership) NextRoundAddOn(inv invoice.Invoice) addon.AddOn {
	return m.AddOn.Plus(addon.New(inv.Tier, inv.TotalDays()))
}

// CarryOverInvoice creates a new invoice based on remaining days of current membership.
// This should only be used when user is upgrading from standard to premium using one-time purchase,
// or switch from one-time purchase to subscription mode.
func (m Membership) CarryOverInvoice() invoice.Invoice {
	return invoice.Invoice{
		ID:         ids.InvoiceID(),
		CompoundID: m.CompoundID,
		Edition:    m.Edition,
		YearMonthDay: dt.YearMonthDay{
			Days: m.RemainingDays(),
		},
		AddOnSource:    addon.SourceCarryOver,
		AppleTxID:      null.String{},
		OrderID:        null.String{},
		OrderKind:      enum.OrderKindAddOn, // All carry-over invoice are add-ons
		PaidAmount:     0,
		PaymentMethod:  m.PaymentMethod,
		StripeSubsID:   null.String{},
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{}, // Will be consumed in the future.
		TimeSlot:       dt.TimeSlot{},
		CarriedOverUtc: chrono.Time{},
	}
}

// ShouldUseAddOn checks if current membership is eligible to
// start using addon.
// If the membership is a zero value, stop.
// If the membership is not expired yet, stop.
func (m Membership) ShouldUseAddOn() error {
	if m.IsZero() {
		return errors.New("stored subscription time can only be claimed by an existing membership")
	}

	if !m.IsExpired() {
		return errors.New("reserved subscription time only comes into effect after current membership expired")
	}

	return nil
}

func (m Membership) HasAddOn() bool {
	return m.AddOn.Standard > 0 || m.AddOn.Premium > 0
}

// withAddOnInvoice transfers invoice to expiration date.
func (m Membership) withAddOnInvoice(i invoice.Invoice) (Membership, error) {
	if !i.IsAddOn() {
		return Membership{}, errors.New("cannot use non-addon invoice as add-on")
	}

	if !i.IsConsumed() {
		return Membership{}, errors.New("invoice not finalized")
	}

	return Membership{
		UserIDs:       m.UserIDs,
		Edition:       i.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(i.EndUTC.Time),
		PaymentMethod: i.PaymentMethod,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		AddOn:         m.AddOn.Clear(i.Tier),
	}.Sync(), nil
}

// addonToInvoice uses a virtual invoice to collect and generate the data used by
// withAddOnInvoice when using AddOn field as a fallback.
func (m Membership) addonToInvoice() invoice.Invoice {
	var days int64
	var payMethod enum.PayMethod
	var tier enum.Tier
	startTime := dt.PickLater(time.Now(), m.ExpireDate.Time)
	var endTime time.Time

	if m.AddOn.Premium != 0 {
		days = m.AddOn.Premium
		endTime = startTime.AddDate(0, 0, int(m.AddOn.Premium))
		tier = enum.TierPremium
	} else if m.AddOn.Standard != 0 {
		days = m.AddOn.Standard
		endTime = startTime.AddDate(0, 0, int(m.AddOn.Standard))
		tier = enum.TierStandard
	}

	if (m.PaymentMethod != enum.PayMethodAli) && (m.PaymentMethod != enum.PayMethodWx) {
		payMethod = enum.PayMethodAli
	} else {
		payMethod = m.PaymentMethod
	}

	return invoice.Invoice{
		ID:         ids.InvoiceID(),
		CompoundID: m.CompoundID,
		Edition: price.Edition{
			Tier:  tier,
			Cycle: enum.CycleYear,
		},
		YearMonthDay: dt.YearMonthDay{
			Days: days,
		},
		AddOnSource:   addon.SourceCarryOver,
		OrderID:       null.String{},
		OrderKind:     enum.OrderKindAddOn,
		PaidAmount:    0,
		PaymentMethod: payMethod,
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.TimeNow(),
		TimeSlot: dt.TimeSlot{
			StartUTC: chrono.TimeFrom(startTime),
			EndUTC:   chrono.TimeFrom(endTime),
		},
	}
}

// pickConsumableAddOn checks if Membership.AddOn is out of sync with invoices.
// In case there are invoices failed to ba saved but membershi's addon field changed,
// use the addon days directly; otherwise we calculate expiration date from invoices.
func (m Membership) pickConsumableAddOn(groupedInv invoice.AddOnGroup) []invoice.Invoice {
	realAddOn := groupedInv.ToAddOn()

	// Use premium first.
	if m.AddOn.Premium != 0 || realAddOn.Premium != 0 {
		consumed := groupedInv.Consume(enum.TierPremium, dt.PickLater(time.Now(), m.ExpireDate.Time))
		// The days saved in membership is larger than those calculated from invoices.
		// Turn the addon days into a virtual invoice and append it to the end of consumed,
		// which will be used to extend membership expiration date.
		if m.AddOn.Premium > realAddOn.Premium {
			consumed = append(consumed, m.addonToInvoice())
		}

		return consumed
	}

	if m.AddOn.Standard != 0 || realAddOn.Standard != 0 {
		consumed := groupedInv.Consume(enum.TierStandard, dt.PickLater(time.Now(), m.ExpireDate.Time))
		if m.AddOn.Standard > realAddOn.Standard {
			consumed = append(consumed, m.addonToInvoice())
		}

		return consumed
	}

	return nil
}

// ClaimAddOns extends expiration date from existing addon invoices, or
// from the AddOn fields if invoices are empty, or the latest invoice's end date
// is less than AddOn.
func (m Membership) ClaimAddOns(inv []invoice.Invoice) (AddOnClaimed, error) {

	// Find out which group of addon could be consumed.
	addOns := m.pickConsumableAddOn(invoice.NewAddOnGroup(inv))

	if addOns == nil || len(addOns) == 0 {
		return AddOnClaimed{}, errors.New("no addon invoice found")
	}

	latest := addOns[len(addOns)-1]
	newM, err := m.withAddOnInvoice(latest)
	if err != nil {
		return AddOnClaimed{}, err
	}

	return AddOnClaimed{
		Invoices:   addOns,
		Membership: newM,
		//Versioned:  newM.Version(NewOrderArchiver(enum.OrderKindAddOn)).WithPriorVersion(m),
		Versioned: NewMembershipVersioned(newM).
			WithPriorVersion(m).
			ArchivedBy(NewArchiver().ByFtcOrder().ActionClaimAddOn()),
	}, nil
}
