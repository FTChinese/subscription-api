package paywall

import (
	"errors"
	"time"

	"github.com/FTChinese/go-rest/rand"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// GenerateMembershipIndex generates a random string to membership id.
func GenerateMembershipIndex() string {
	return "mmb_" + rand.String(12)
}

// Membership contains a user's membership details
// This is actually called subscription by Stripe.
// TODO: rename ID to avoid conflict when embedded.
type Membership struct {
	ID null.String `json:"id" db:"sub_id"` // A random string. Not used yet.
	reader.MemberID
	LegacyTier   null.Int `json:"-" db:"vip_type"`
	LegacyExpire null.Int `json:"-" db:"expire_time"`
	Coordinate
	ExpireDate    chrono.Date    `json:"expireDate" db:"sub_expire_date"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"sub_pay_method"`
	StripeSubID   null.String    `json:"-" db:"stripe_sub_id"`
	StripePlanID  null.String    `json:"-" db:"stripe_plan_id"`
	AutoRenewal   bool           `json:"autoRenewal" db:"sub_auto_renew"`
	// This is used to save stripe subscription status.
	// Since wechat and alipay treats everything as one-time purchase, they do not have a complex state machine.
	// If we could integrate apple in-app purchase, this column
	// might be extended to apple users.
	// Only `active` should be treated as valid member.
	// Wechat and alipay defaults to `active` for backward compatibility.
	Status SubStatus `json:"status" db:"sub_status"`
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(accountID reader.MemberID) Membership {
	return Membership{
		ID:       null.StringFrom(GenerateMembershipIndex()),
		MemberID: accountID,
	}
}

// GenerateID generates a unique id for this membership if
// it is not set. This id is mostly used to identify a row
// in restful api.
func (m *Membership) GenerateID() {
	if m.ID.Valid {
		return
	}

	m.ID = null.StringFrom(GenerateMembershipIndex())
}

// Deprecate
func (m Membership) TierCode() int64 {
	switch m.Tier {
	case enum.TierStandard:
		return 10
	case enum.TierPremium:
		return 100
	}

	return 0
}

// Normalize turns legacy vip_type and expire_time into
// member_tier and expire_date columns, or vice versus.
func (m *Membership) Normalize() {
	// Turn unix seconds to time.
	if m.LegacyExpire.Valid && m.ExpireDate.IsZero() {
		m.ExpireDate = chrono.DateFrom(time.Unix(m.LegacyExpire.Int64, 0))
	}

	// Turn time to unix seconds.
	if !m.ExpireDate.IsZero() && m.LegacyExpire.IsZero() {
		m.LegacyExpire = null.IntFrom(m.ExpireDate.Unix())
	}

	if m.LegacyTier.Valid && m.Tier == enum.InvalidTier {
		switch m.LegacyTier.Int64 {
		case 10:
			m.Tier = enum.TierStandard
		case 100:
			m.Tier = enum.TierPremium
		}
	}

	if m.Tier != enum.InvalidTier && m.LegacyTier.IsZero() {
		switch m.Tier {
		case enum.TierStandard:
			m.LegacyTier = null.IntFrom(10)
		case enum.TierPremium:
			m.LegacyTier = null.IntFrom(100)
		}
	}
}

// IsZero test whether the instance is empty.
func (m Membership) IsZero() bool {
	return m.CompoundID == "" && m.Tier == enum.InvalidTier
}

// NewStripe creates a new membership for stripe.
func (m Membership) NewStripe(id reader.MemberID, p StripeSubParams, s *stripe.Subscription) Membership {

	m.GenerateID()

	periodEnd := canonicalizeUnix(s.CurrentPeriodEnd)

	status, _ := ParseSubStatus(string(s.Status))

	return Membership{
		ID:       m.ID,
		MemberID: id,
		Coordinate: Coordinate{
			Tier:  p.Tier,
			Cycle: p.Cycle,
		},
		ExpireDate:    chrono.DateFrom(periodEnd.AddDate(0, 0, 1)),
		PaymentMethod: enum.PayMethodStripe,
		StripeSubID:   null.StringFrom(s.ID),
		StripePlanID:  null.StringFrom(p.GetStripePlanID()),
		AutoRenewal:   !s.CancelAtPeriodEnd,
		Status:        status,
	}
}

// WithStripe update an existing stripe membership.
// This is used in webhook.
func (m Membership) WithStripe(id reader.MemberID, s *stripe.Subscription) (Membership, error) {

	m.GenerateID()

	periodEnd := canonicalizeUnix(s.CurrentPeriodEnd)

	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.AutoRenewal = !s.CancelAtPeriodEnd
	m.Status, _ = ParseSubStatus(string(s.Status))

	return m, nil
}

// FromAliOrWx builds a new membership based on a confirmed
// order.
func (m Membership) FromAliOrWx(sub Order) (Membership, error) {
	if !sub.IsConfirmed() {
		return m, errors.New("only confirmed order could be used to build membership")
	}

	m.GenerateID()

	if m.IsZero() {
		m.CompoundID = sub.CompoundID
		m.FtcID = sub.FtcID
		m.UnionID = sub.UnionID
	}

	m.Tier = sub.Tier
	m.Cycle = sub.Cycle
	m.ExpireDate = sub.EndDate
	m.PaymentMethod = sub.PaymentMethod
	m.StripeSubID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = false

	return m, nil
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

func (m Membership) IsAliOrWxPay() bool {
	if m.Tier != enum.InvalidTier && m.PaymentMethod == enum.InvalidPay {
		return true
	}

	return m.PaymentMethod == enum.PayMethodAli || m.PaymentMethod == enum.PayMethodWx
}

// IsExpired tests if the membership's expiration date is before now.
func (m Membership) IsExpired() bool {
	// If membership does not exist, it is treated as expired.
	if m.IsZero() {
		return true
	}

	// If expire date is before now, AND auto renew is false,
	// we treat this one as actually expired.
	// If ExpireDate is passed, but auto renew is true, we still
	// treat this one as not expired.
	return m.ExpireDate.Before(time.Now().Truncate(24*time.Hour)) && !m.AutoRenewal
}

// CanRenew tests if a membership is allowed to renew subscription.
// A member could only renew its subscription when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renewal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
// Algorithm changed to membership duration not larger than 3 years.

func (m Membership) inRenewalPeriod() bool {
	today := time.Now().Truncate(24 * time.Hour)
	threeYearsLater := today.AddDate(3, 0, 0)

	// If today is after expiration date, it means the membership
	// is already expired.
	// expire date >= today
	if today.After(m.ExpireDate.Time) {
		return false
	}

	// expire date <= three years later
	if threeYearsLater.Before(m.ExpireDate.Time) {
		return false
	}

	return true
}

// PermitRenewal test if current membership is allowed to renew.
// now ---------3 years ---------> Expire date
// expire date - now <= 3 years
func (m Membership) PermitRenewal() bool {
	if m.ExpireDate.IsZero() {
		return false
	}

	if m.AutoRenewal {
		return false
	}

	if !m.IsAliOrWxPay() {
		return false
	}

	return m.inRenewalPeriod()
}

func (m Membership) IsValidPremium() bool {
	return m.Tier == enum.TierPremium && !m.IsExpired()
}

// SubsKind determines what kind of order a user is creating.
func (m Membership) SubsKind(p Plan) (SubsKind, error) {
	if m.IsZero() {
		return SubsKindCreate, nil
	}

	if m.Status != SubStatusNull && m.Status.ShouldCreate() {
		return SubsKindCreate, nil
	}

	if m.IsExpired() {
		return SubsKindCreate, nil
	}

	// Renewal.
	if m.Tier == p.Tier {
		if m.inRenewalPeriod() {
			return SubsKindRenew, nil
		} else {
			return SubsKindDeny, errors.New("current membership expiration date is beyond max renewal period")
		}
	}

	if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
		return SubsKindUpgrade, nil
	}

	return SubsKindDeny, errors.New("unknown subscription usage")
}

// PermitStripeCreate checks whether subscription
// via stripe is permitted or not.
// Cases for permission:
// 1. Membership does not exist;
// 2. Membership exists via alipay or wxpay but expired;
// 3. Membership exits via stripe but is expired and it is not auto renewable.
// 4. A stripe subscription that is not expired, auto renewable, but its current status is not active.
// Returns errors indicates permission not allowed and gives reason:
// 1. ErrNonStripeValidSub - a valid subscription not paid via stripe
// 2. ErrActiveStripeSub - a valid stripe subscription.
// 3. ErrUnknownSubState
func (m Membership) PermitStripeCreate() error {
	// If a membership does not exist, allow create stripe
	// subscription
	if m.IsZero() {
		return nil
	}

	if m.IsAliOrWxPay() {
		if m.IsExpired() {
			return nil
		}
		return util.ErrNonStripeValidSub
	}

	if m.PaymentMethod == enum.PayMethodStripe {
		// An expired member that is not auto renewal.
		if m.IsExpired() && !m.AutoRenewal {
			return nil
		}
		// Member is not expired, or is auto renewal.
		// Deny such request.
		// If status is active, deny it.
		if m.Status.ShouldCreate() {
			return nil
		}

		// Now it is not expired, or auto renewal,
		// and status is active.
		return util.ErrActiveStripeSub
	}

	// Member is either not expired, or auto renewal
	// Deny any other cases.
	return util.ErrUnknownSubState
}

// PermitStripeUpgrade tests whether a stripe customer with
// standard membership should be allowed to upgrade to premium.
func (m Membership) PermitStripeUpgrade() bool {
	if m.IsZero() {
		return false
	}

	if m.PaymentMethod != enum.PayMethodStripe {
		return false
	}

	if m.IsExpired() {
		return false
	}

	// Membership is not expired, or is auto renewable, but status is not active.
	if m.Status != SubStatusActive {
		return false
	}

	return m.Tier == enum.TierStandard
}
