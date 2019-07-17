package paywall

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/stripe/stripe-go"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func genMmID() (string, error) {
	s, err := gorest.RandomBase64(9)
	if err != nil {
		return "", err
	}

	return "mmb_" + s, nil
}

// Membership contains a user's membership details
// This is actually called subscription by Stripe.
type Membership struct {
	ID            null.String    `json:"id"`
	CompoundID    string         `json:"-"` // Either FTCUserID or UnionID
	FTCUserID     null.String    `json:"-"`
	UnionID       null.String    `json:"-"` // For both vip_id_alias and wx_union_id columns.
	Tier          enum.Tier      `json:"tier"`
	Cycle         enum.Cycle     `json:"billingCycle"`
	ExpireDate    chrono.Date    `json:"expireDate"`
	PaymentMethod enum.PayMethod `json:"payMethod"`
	StripeSubID   null.String    `json:"-"`
	StripePlanID  null.String    `json:"-"`
	AutoRenewal   bool           `json:"autoRenewal"`
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(u UserID) Membership {
	id, _ := genMmID()
	return Membership{
		ID:         null.StringFrom(id),
		CompoundID: u.CompoundID,
		FTCUserID:  u.FtcID,
		UnionID:    u.UnionID,
	}
}

// FromStripe creates a new Membership purchased via stripe.
// The original might not exists, which indicates this is a new member.
// Even if we only allow FTC user to use Stripe, we still
// needs to record incomming user's wechat id, since this user
// might already have its accounts linked.
func (m Membership) FromStripe(
	id UserID,
	p StripeSubParams,
	sub *stripe.Subscription) Membership {

	endTime := time.Unix(sub.CurrentPeriodEnd, 0)

	if m.ID.IsZero() {
		mId, _ := genMmID()
		m.ID = null.StringFrom(mId)
	}

	if m.IsZero() {
		return Membership{
			ID:            m.ID,
			CompoundID:    id.CompoundID,
			FTCUserID:     id.FtcID,
			UnionID:       id.UnionID,
			Tier:          p.Tier,
			Cycle:         p.Cycle,
			ExpireDate:    chrono.DateFrom(endTime.AddDate(0, 0, 1)),
			PaymentMethod: enum.PayMethodStripe,
			StripeSubID:   null.StringFrom(sub.ID),
			StripePlanID:  null.StringFrom(p.PlanID),
			AutoRenewal:   !sub.CancelAtPeriodEnd,
		}
	}

	return Membership{
		ID:            m.ID,
		CompoundID:    m.CompoundID,
		FTCUserID:     m.FTCUserID,
		UnionID:       m.UnionID,
		Tier:          p.Tier,
		Cycle:         p.Cycle,
		ExpireDate:    chrono.DateFrom(endTime.AddDate(0, 0, 1)),
		PaymentMethod: enum.PayMethodStripe,
		StripeSubID:   null.StringFrom(sub.ID),
		StripePlanID:  null.StringFrom(p.PlanID),
		AutoRenewal:   !sub.CancelAtPeriodEnd,
	}
}

func (m Membership) RefreshStripe(s *stripe.Subscription) Membership {

	planID, err := extractStripePlanID(s)
	if err == nil {
		m.StripePlanID = null.StringFrom(planID)
	}

	//m.Tier =          p.Tier
	//m.Cycle =         p.Cycle
	m.ExpireDate = chrono.DateFrom(time.Unix(s.CurrentPeriodEnd, 0))
	m.PaymentMethod = enum.PayMethodStripe
	m.StripeSubID = null.StringFrom(s.ID)
	m.AutoRenewal = !s.CancelAtPeriodEnd

	return m
}

func (m Membership) FromAliOrWx(sub Subscription) (Membership, error) {
	if !sub.IsConfirmed {
		return m, errors.New("only confirmed order could be used to build membership")
	}

	if m.ID.IsZero() {
		mId, _ := genMmID()
		m.ID = null.StringFrom(mId)
	}

	if m.IsZero() {
		return Membership{
			ID:            m.ID,
			CompoundID:    sub.CompoundID,
			FTCUserID:     sub.FtcID,
			UnionID:       sub.UnionID,
			Tier:          sub.TierToBuy,
			Cycle:         sub.BillingCycle,
			ExpireDate:    sub.EndDate,
			PaymentMethod: sub.PaymentMethod,
			StripeSubID:   null.String{},
			AutoRenewal:   false,
		}, nil
	}

	return Membership{
		ID:            m.ID,
		CompoundID:    m.CompoundID,
		FTCUserID:     m.FTCUserID,
		UnionID:       m.UnionID,
		Tier:          sub.TierToBuy,
		Cycle:         sub.BillingCycle,
		ExpireDate:    sub.EndDate,
		PaymentMethod: sub.PaymentMethod,
		StripeSubID:   null.String{},
		AutoRenewal:   false,
	}, nil
}

// FromGiftCard creates a new instance based on a gift card.
func (m Membership) FromGiftCard(c GiftCard) (Membership, error) {

	var expTime time.Time

	expTime, err := c.ExpireTime()

	if err != nil {
		return m, err
	}

	m.Tier = c.Tier
	m.Cycle = c.CycleUnit
	m.ExpireDate = chrono.DateFrom(expTime)

	return m, nil
}

func (m Membership) Exists() bool {
	return m.CompoundID != "" && m.Tier != enum.InvalidTier && m.Cycle != enum.InvalidCycle
}

func (m Membership) IsFtc() bool {
	return m.FTCUserID.Valid
}

func (m Membership) IsWx() bool {
	return m.UnionID.Valid
}

// CanRenew tests if a membership is allowed to renew subscription.
// A member could only renew its subscription when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renewal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
// Algorithm changed to membership duration not larger than 3 years.
// Deprecate
//func (m Membership) CanRenew(cycle enum.Cycle) bool {
//	cycleEnds, err := cycle.TimeAfterACycle(time.Now())
//
//	if err != nil {
//		return false
//	}
//
//	return m.ExpireDate.Before(cycleEnds)
//}

// IsRenewAllowed test if current membership is allowed to renew.
// now ---------3 years ---------> Expire date
// expire date - now <= 3 years
func (m Membership) IsRenewAllowed() bool {
	return m.ExpireDate.Before(time.Now().AddDate(3, 0, 0))
}

// IsExpired tests if the membership's expiration date is before now.
func (m Membership) IsExpired() bool {
	if m.ExpireDate.IsZero() {
		return true
	}
	// If expire is before now, it is expired.
	return m.ExpireDate.Before(time.Now().Truncate(24 * time.Hour))
}

// SubsKind determines what kind of order a user is creating.
func (m Membership) SubsKind(p Plan) (SubsKind, error) {
	if m.IsZero() {
		return SubsKindCreate, nil
	}

	// If member is expired.
	if m.IsExpired() {
		return SubsKindCreate, nil
	}

	// If member duration is beyond renewal period.
	if !m.IsRenewAllowed() {
		return SubsKindDeny, ErrBeyondRenewal
	}

	if m.Tier == p.Tier {
		return SubsKindRenew, nil
	}

	if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
		return SubsKindUpgrade, nil
	}

	// The only possibility left here is:
	// m.Tier == enum.Premium && p.Tier = enum.TierStandard
	return SubsKindDeny, ErrDowngrade
}

func (m Membership) IsZero() bool {
	return m.Tier == enum.InvalidTier
}

func (m Membership) IsAliOrWxPay() bool {
	return m.PaymentMethod == enum.InvalidPay || m.PaymentMethod == enum.PayMethodAli || m.PaymentMethod == enum.PayMethodWx
}

func (m Membership) ActionOnStripe() StripeAction {
	if m.Tier == enum.InvalidTier {
		return StripeActionCreate
	}

	// if membership is not expired yet, do nothing.
	if !m.IsExpired() {
		return StripeActionNoop
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx, enum.InvalidPay:
		return StripeActionCreate

	case enum.PayMethodStripe:
		if m.AutoRenewal {
			return StripeActionSync
		} else {
			return StripeActionCreate
		}

	default:
		return StripeActionDeny
	}
}

func (m Membership) PermitStripeCreate() bool {
	// If a membership does not exist, allow create stripe
	// subscription
	if m.IsZero() {
		return true
	}

	// If a membership is expired but not auto renewal,
	// allow create subscription.
	// This includes non-existent member.
	if m.IsExpired() && !m.AutoRenewal {
		return true
	}

	// Deny any other cases.
	return false
}

func (m Membership) PermitStripeUpgrade(p StripeSubParams) bool {
	if m.PaymentMethod != enum.PayMethodStripe {
		return false
	}

	if m.IsExpired() {
		return false
	}

	if m.Tier == p.Tier && m.Cycle == m.Cycle {
		return false
	}

	if m.Tier >= p.Tier {
		return false
	}

	if m.Cycle >= p.Cycle {
		return false
	}

	return true
}

type StripeAction int

const (
	StripeActionDeny StripeAction = iota
	StripeActionNoop
	StripeActionCreate
	StripeActionSync
)
