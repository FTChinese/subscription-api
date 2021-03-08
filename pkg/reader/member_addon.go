package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/guregu/null"
	"time"
)

// AddOnClaimed contains the data that should be
// used to perform add-on transfer.
type AddOnClaimed struct {
	Invoices   []invoice.Invoice // The invoices used to extend membership's expiration date. They will not be used next time.
	Membership Membership        // Update membership.
	Snapshot   MemberSnapshot
}

func (m Membership) PlusAddOn(addOn addon.AddOn) Membership {
	m.AddOn = m.AddOn.Plus(addOn)
	return m
}

func (m Membership) CarriedOverAddOn() addon.AddOn {
	return addon.New(m.Tier, m.RemainingDays())
}

// CarryOverInvoice creates a new invoice based on remainig days of current membership.
// This should only be used when user is upgrading from standard to premium using one-time purchase,
// or switch from one-=time purchase to subscription mode.
func (m Membership) CarryOverInvoice() invoice.Invoice {
	return invoice.Invoice{
		ID:         pkg.InvoiceID(),
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
		PriceID:        m.FtcPlanID,
		StripeSubsID:   null.String{},
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{}, // Will be consumed in the future.
		DateTimePeriod: dt.DateTimePeriod{},
		CarriedOverUtc: chrono.Time{},
	}
}

func (m Membership) ShouldUseAddOn() error {
	if m.IsZero() {
		return errors.New("reserved subscription time only be claimed by an existing membership")
	}

	if !m.IsExpired() {
		return errors.New("reserved subscription time only comes into effect after current membership expired")
	}

	return nil
}

func (m Membership) claimAddOn(i invoice.Invoice) (Membership, error) {
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
		FtcPlanID:     i.PriceID,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		AddOn:         m.AddOn.Clear(i.Tier),
	}.Sync(), nil
}

func (m Membership) ClaimAddOns(inv []invoice.Invoice) (AddOnClaimed, error) {

	addOns := invoice.NewAddOnGroup(inv).
		Consumable(dt.PickLater(time.Now(), m.ExpireDate.Time))

	if len(addOns) == 0 {
		return AddOnClaimed{}, errors.New("no addon invoice found")
	}

	latest := addOns[len(addOns)-1]
	newM, err := m.claimAddOn(latest)
	if err != nil {
		return AddOnClaimed{}, err
	}

	return AddOnClaimed{
		Invoices:   addOns,
		Membership: newM,
		Snapshot:   m.Snapshot(FtcArchiver(enum.OrderKindAddOn)),
	}, nil
}
