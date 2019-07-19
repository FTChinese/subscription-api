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
	ID null.String `json:"id"`
	//CompoundID    string         `json:"-"` // Either FtcID or UnionID
	//FtcID         null.String    `json:"-"`
	//UnionID       null.String    `json:"-"` // For both vip_id_alias and wx_union_id columns.
	UserID
	Tier          enum.Tier      `json:"tier"`
	Cycle         enum.Cycle     `json:"cycle"`
	ExpireDate    chrono.Date    `json:"expireDate"`
	PaymentMethod enum.PayMethod `json:"payMethod"`
	StripeSubID   null.String    `json:"-"`
	StripePlanID  null.String    `json:"-"`
	AutoRenewal   bool           `json:"autoRenewal"`
	// This is used to save stripe subscription status.
	// Since wechat and alipay treats everything as one-time purchase, they do not have a complex state machine.
	// If we could integrate apple in-app purchase, this column
	// might be extended to apple users.
	// Only `active` should be treated as valid member.
	// Wechat and alipay default to `active` for backward compatibility.
	Status SubStatus `json:"status"`
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(u UserID) Membership {
	id, _ := genMmID()
	return Membership{
		ID: null.StringFrom(id),
		//CompoundID: u.CompoundID,
		//FtcID:      u.FtcID,
		//UnionID:    u.UnionID,
		UserID: u,
	}
}

// FromStripe creates a new Membership purchased via stripe.
// The original might not exists, which indicates this is a new member.
// Even if we only allow FTC user to use Stripe, we still
// needs to record incomming user's wechat id, since this user
// might already have its accounts linked.
func (m Membership) FromStripe(
	id UserID,
	sub *stripe.Subscription) (Membership, error) {

	// Must test before modifying data.
	if m.IsZero() {
		m.CompoundID = id.CompoundID
		m.FtcID = id.FtcID
		m.UnionID = id.UnionID

		mId, _ := genMmID()
		m.ID = null.StringFrom(mId)
	}

	planID, err := extractStripePlanID(sub)
	if err != nil {
		return m, err
	}
	m.StripePlanID = null.StringFrom(planID)

	plan, err := GetStripeToFtcPlans(sub.Livemode).GetPlanByID(planID)
	if err != nil {
		return m, err
	}

	m.Tier = plan.Tier
	m.Cycle = plan.Cycle
	m.ExpireDate = chrono.DateFrom(time.Unix(sub.CurrentPeriodEnd, 0).AddDate(0, 0, 1))
	m.PaymentMethod = enum.PayMethodStripe
	m.StripeSubID = null.StringFrom(sub.ID)
	m.AutoRenewal = !sub.CancelAtPeriodEnd
	m.Status, _ = ParseSubStatus(string(sub.Status))

	return m, nil
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
			ID: m.ID,
			UserID: UserID{
				CompoundID: sub.CompoundID,
				FtcID:      sub.FtcID,
				UnionID:    sub.UnionID,
			},
			Tier:          sub.TierToBuy,
			Cycle:         sub.BillingCycle,
			ExpireDate:    sub.EndDate,
			PaymentMethod: sub.PaymentMethod,
			StripeSubID:   null.String{},
			AutoRenewal:   false,
		}, nil
	}

	return Membership{
		ID: m.ID,
		UserID: UserID{
			CompoundID: m.CompoundID,
			FtcID:      m.FtcID,
			UnionID:    m.UnionID,
		},
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
	return m.FtcID.Valid
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

func (m Membership) PermitStripeCreate() bool {
	// If a membership does not exist, allow create stripe
	// subscription
	if m.IsZero() {
		return true
	}

	if m.IsAliOrWxPay() {
		if m.IsExpired() {
			return true
		}
		return false
	}

	if m.PaymentMethod == enum.PayMethodStripe {
		// An expired member that is not auto renewal.
		if m.IsExpired() && !m.AutoRenewal {
			return true
		}
		// Member is not expired, or is auto renewal.
		if m.Status == SubStatusActive {
			return false
		}

		return false
	}

	// Member is either not expired, or auto renewal
	// Deny any other cases.
	return false
}

// PermitStripeUpgrade tests whether a stripe customer with
// standard membership should be allowed to upgrade to premium.
func (m Membership) PermitStripeUpgrade(p StripeSubParams) bool {
	// expired members could simply re-subscribe.
	// If member is expired but it is an auto renewal, we should allow upgrading.
	if m.PaymentMethod == enum.PayMethodStripe && m.Tier < p.Tier {
		if m.IsExpired() && !m.AutoRenewal {
			return false
		}

		return true
	}

	return false
}
