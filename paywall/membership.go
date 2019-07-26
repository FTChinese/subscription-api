package paywall

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"gitlab.com/ftchinese/subscription-api/util"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func GenerateMemberID() (string, error) {
	s, err := gorest.RandomBase64(9)
	if err != nil {
		return "", err
	}

	return "mmb_" + s, nil
}

// Membership contains a user's membership details
// This is actually called subscription by Stripe.
type Membership struct {
	ID null.String `json:"id"` // A random string. Not used yet.
	//AccountID
	User AccountID
	Coordinate
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
	// Wechat and alipay defaults to `active` for backward compatibility.
	Status SubStatus `json:"status"`
	//FtcPlan Plan      `json:"-"` // The subscribed plan attached to this membership.
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(accountID AccountID) Membership {
	id, _ := GenerateMemberID()
	return Membership{
		ID:   null.StringFrom(id),
		User: accountID,
	}
}

func (m Membership) TierCode() int64 {
	switch m.Tier {
	case enum.TierStandard:
		return 10
	case enum.TierPremium:
		return 100
	}

	return 0
}

// FromStripe creates a new Membership purchased via stripe.
// The original might not exists, which indicates this is a new member.
// Even if we only allow FTC user to use Stripe, we still
// needs to record incomming user's wechat id, since this user
// might already have its accounts linked.
func (m Membership) FromStripe(
	id AccountID,
	sub StripeSub) (Membership, error) {

	if m.ID.IsZero() {
		mId, _ := GenerateMemberID()
		m.ID = null.StringFrom(mId)
	}

	// Must test before modifying data.
	if m.IsZero() {
		m.User = id
	}

	plan, err := sub.BuildFtcPlan()
	if err != nil {
		return m, err
	}

	m.Tier = plan.Tier
	m.Cycle = plan.Cycle
	m.ExpireDate = chrono.DateFrom(sub.CurrentPeriodEnd.AddDate(0, 0, 1))
	m.PaymentMethod = enum.PayMethodStripe
	m.StripeSubID = null.StringFrom(sub.ID)
	m.StripePlanID = null.StringFrom(sub.Plan.ID)
	m.AutoRenewal = !sub.CancelAtPeriodEnd
	m.Status, _ = ParseSubStatus(string(sub.Status))

	return m, nil
}

func (m Membership) FromAliOrWx(sub Subscription) (Membership, error) {
	if !sub.IsConfirmed {
		return m, errors.New("only confirmed order could be used to build membership")
	}

	if m.ID.IsZero() {
		mId, _ := GenerateMemberID()
		m.ID = null.StringFrom(mId)
	}

	if m.IsZero() {
		m.User = sub.User
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

// CanRenew tests if a membership is allowed to renew subscription.
// A member could only renew its subscription when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renewal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
// Algorithm changed to membership duration not larger than 3 years.

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

func (m Membership) IsValidPremium() bool {
	return m.Tier == enum.TierPremium && !m.IsExpired()
}

func (m Membership) IsZero() bool {
	return m.Tier == enum.InvalidTier
}

func (m Membership) IsAliOrWxPay() bool {
	return m.PaymentMethod == enum.InvalidPay || m.PaymentMethod == enum.PayMethodAli || m.PaymentMethod == enum.PayMethodWx
}

// SubsKind determines what kind of order a user is creating.
func (m Membership) SubsKind(p Plan) (SubsKind, error) {
	if m.IsZero() {
		return SubsKindCreate, nil
	}

	if m.IsAliOrWxPay() {
		if m.IsExpired() {
			return SubsKindCreate, nil
		}

		if !m.IsRenewAllowed() {
			return SubsKindDeny, util.ErrBeyondRenewal
		}

		if m.Tier == p.Tier {
			return SubsKindRenew, nil
		}

		if m.Tier < p.Tier {
			return SubsKindUpgrade, nil
		}

		return SubsKindDeny, util.ErrDowngrade
	}

	return SubsKindDeny, errors.New("not alipay or wechat pay")
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
		// If status is active, deny it.
		if m.Status == SubStatusActive {
			return util.ErrActiveStripeSub
		}

		return nil
	}

	// Member is either not expired, or auto renewal
	// Deny any other cases.
	return util.ErrUnknownSubState
}

// PermitStripeUpgrade tests whether a stripe customer with
// standard membership should be allowed to upgrade to premium.
func (m Membership) PermitStripeUpgrade(p StripeSubParams) bool {
	if m.IsZero() {
		return false
	}

	if m.PaymentMethod != enum.PayMethodStripe {
		return false
	}

	if m.IsExpired() && !m.AutoRenewal {
		return false
	}

	// Membership is not expired, or is auto renewable, but status is not active.
	if m.Status != SubStatusActive {
		return false
	}

	// Already subscribed to premium edition.
	if m.Tier == enum.TierPremium {
		return false
	}

	return m.GetOrdinal() < p.GetOrdinal()
}
