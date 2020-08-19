package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/redeem"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

var tierToCode = map[enum.Tier]int64{
	enum.TierStandard: 10,
	enum.TierPremium:  100,
}

var codeToTier = map[int64]enum.Tier{
	10:  enum.TierStandard,
	100: enum.TierPremium,
}

// Membership contains a user's membership details
// This is actually called subscription by Stripe.
type Membership struct {
	reader.MemberID
	product.Edition
	LegacyTier    null.Int       `json:"-" db:"vip_type"`
	LegacyExpire  null.Int       `json:"-" db:"expire_time"`
	ExpireDate    chrono.Date    `json:"expireDate" db:"expire_date"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"pay_method"`
	StripeSubID   null.String    `json:"-" db:"stripe_subs_id"`
	StripePlanID  null.String    `json:"-" db:"stripe_plan_id"`
	AutoRenew     bool           `json:"autoRenew" db:"sub_auto_renew"`
	// This is used to save stripe subscription status.
	// Since wechat and alipay treats everything as one-time purchase, they do not have a complex state machine.
	// If we could integrate apple in-app purchase, this column
	// might be extended to apple users.
	// Only `active` should be treated as valid member.
	// Wechat and alipay defaults to `active` for backward compatibility.
	Status       enum.SubsStatus `json:"status" db:"subs_status"`
	AppleSubID   null.String     `json:"-" db:"apple_subs_id"`
	B2BLicenceID null.String     `json:"b2bLicenceId" db:"b2b_licence_id"`
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(accountID reader.MemberID) Membership {
	return Membership{
		MemberID: accountID,
	}
}

// IsZero test whether the instance is empty.
func (m Membership) IsZero() bool {
	return m.CompoundID == "" && m.Tier == enum.TierNull
}

// IsExpired tests if the membership's expiration date is before now.
// A non-existing membership is treated as expired.
// Auto renewal is treated as not expired.
func (m Membership) IsExpired() bool {
	// If membership does not exist, it is treated as expired.
	if m.IsZero() {
		return true
	}

	// If expire date is before now, AND auto renew is false,
	// we treat this one as actually expired.
	// If ExpireDate is passed, but auto renew is true, we still
	// treat this one as not expired.
	return m.ExpireDate.Before(time.Now().Truncate(24*time.Hour)) && !m.AutoRenew
}

func (m Membership) IsEqual(other Membership) bool {
	if m.IsZero() && other.IsZero() {
		return true
	}

	return m.CompoundID == other.CompoundID && m.AppleSubID.String == other.AppleSubID.String
}

func (m Membership) IsValid() bool {
	if m.IsZero() {
		return false
	}

	// If it is expired, check whether auto renew is on.
	if m.IsExpired() {
		if !m.AutoRenew {
			return false
		}

		return true
	}

	return true
}

// Normalize turns legacy vip_type and expire_time into
// member_tier and expire_date columns, or vice versus.
func (m *Membership) Normalize() Membership {
	if m.IsZero() {
		return *m
	}

	legacyDate := time.Unix(m.LegacyExpire.Int64, 0)

	// Use whichever comes later.
	// If LegacyExpire is after ExpireDate, then we should
	// use LegacyExpire and LegacyTier
	if legacyDate.After(m.ExpireDate.Time) {
		m.ExpireDate = chrono.DateFrom(legacyDate)
		m.Tier = codeToTier[m.LegacyTier.Int64]
	} else {
		m.LegacyExpire = null.IntFrom(m.ExpireDate.Unix())
		m.LegacyTier = null.IntFrom(tierToCode[m.Tier])
	}

	return *m
}

// IsAliOrWxPay checks whether current membership is purchased
// via alipay or wechat pay.
func (m Membership) IsAliOrWxPay() bool {
	// For backward compatibility. If Tier field comes from LegacyTier, then PayMethod field will be null.
	// We treat all those cases as wxpay or alipay.
	if m.Tier != enum.TierNull && m.PaymentMethod == enum.PayMethodNull {
		return true
	}

	return m.PaymentMethod == enum.PayMethodAli || m.PaymentMethod == enum.PayMethodWx
}

// IsIAP tests whether this membership comes from Apple.
func (m Membership) IsIAP() bool {
	return m.AppleSubID.Valid
}

func (m Membership) IsB2B() bool {
	return m.B2BLicenceID.Valid
}

func (m Membership) IsValidPremium() bool {
	return m.Tier == enum.TierPremium && !m.IsExpired()
}

// inRenewalPeriod tests if a membership is allowed to renew subscription for alipay or wecaht pay.
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

// PermitAliOrWxPay checks whether user is allowed to pay via
// alipay or wechat pay.
// A zero or expired membership permit pay by all means.
// If current membership comes from Stripe, IAP or B2B, deny it.
func (m Membership) PermitAliOrWxPay() bool {
	return m.IsZero() || m.IsExpired() || m.IsAliOrWxPay()
}

// PermitRenewal test if current membership is allowed to renew.
// now ---------3 years ---------> Expire date
// expire date - now <= 3 years
func (m Membership) PermitRenewal() bool {
	if m.ExpireDate.IsZero() {
		return false
	}

	if m.AutoRenew {
		return false
	}

	if !m.IsAliOrWxPay() {
		return false
	}

	return m.inRenewalPeriod()
}

func (m Membership) PermitAliWxUpgrade() bool {
	if m.IsZero() {
		return false
	}

	if !m.IsAliOrWxPay() {
		return false
	}

	if m.IsExpired() {
		return false
	}

	if m.Tier != enum.TierStandard {
		return false
	}

	return true
}

// SubsKind determines what kind of order a user is creating based on existing membership.
// SubsKind   |   Membership
// ---------------------------
// Create     |   Zero / Status is not active / Expired
// Renewal    |   Tier === Plan.Tier
// Upgrade    |   Tier is Standard while Plan.Tier is Premium
func (m Membership) SubsKind(p product.ExpandedPlan) (enum.OrderKind, error) {
	if m.IsZero() {
		return enum.OrderKindCreate, nil
	}

	if m.Status != enum.SubsStatusNull && m.Status.ShouldCreate() {
		return enum.OrderKindCreate, nil
	}

	if m.IsExpired() {
		return enum.OrderKindCreate, nil
	}

	// Renewal.
	if m.Tier == p.Tier {
		if m.inRenewalPeriod() {
			return enum.OrderKindRenew, nil
		} else {
			return enum.OrderKindNull, ErrRenewalForbidden
		}
	}

	if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
		return enum.OrderKindUpgrade, nil
	}

	if m.Tier == enum.TierPremium && p.Tier == enum.TierStandard {
		return enum.OrderKindNull, ErrDowngradeForbidden
	}

	return enum.OrderKindNull, ErrUnknownSubsKind
}

// FromAliWxOrder build/create a new membership based on an confirmed order.
func (m Membership) FromAliWxOrder(order Order) (Membership, error) {
	if !order.IsConfirmed() {
		return m, fmt.Errorf("payment order %s is not confirmed yet", order.ID)
	}

	if m.IsZero() {
		m.MemberID = order.MemberID
	}

	m.Tier = order.Tier
	m.Cycle = order.Cycle
	m.ExpireDate = order.EndDate
	m.PaymentMethod = order.PaymentMethod
	m.StripeSubID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenew = false
	m.AppleSubID = null.String{}
	m.B2BLicenceID = null.String{}

	return m, nil
}

// ================================
// The following section handles stripe.
// ================================

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
		return ErrNonStripeValidSub
	}

	if m.PaymentMethod == enum.PayMethodStripe {
		// An expired member that is not auto renewal.
		if m.IsExpired() && !m.AutoRenew {
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
		return ErrActiveStripeSub
	}

	// Member is either not expired, or auto renewal
	// Deny any other cases.
	return ErrUnknownSubState
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
	if m.Status != enum.SubsStatusActive {
		return false
	}

	return m.Tier == enum.TierStandard
}

// =======================================
// The following section handles Apple IAP
// =======================================

// MergeIAPMembership merges iap membership into an FTC membership.
// Only two cases are allowed to merge:
// * Both sides are refer to the same membership (including zero value);
// * IAP side is zero and FTC side non-zero but invalid.
// Let's imagine there are two numbers: a and b
func (m Membership) MergeIAPMembership(iapMember Membership) (Membership, error) {
	// a == b, a and b could be both 0 or non-0.
	// Equal means either both are zero values, or they
	// refer to the same instance.
	// In such case it is fine to return any of them.
	// The caller should then check whether the returned value
	// is zero.
	// For zero value, build a new membership based on IAP
	// transaction; otherwise just update it.
	if iapMember.IsEqual(m) {
		return m, nil
	}

	// a != b:
	// a == 0, b != 0;
	// a != 0, b == 0;
	// a != 0, b != 0.
	// The two sides are not equal. They must be totally
	// different memberships.
	// If the IAP side is non-zero, this means it is
	// already linked to an FTC account and now  is trying
	// to link to another FTC  account which should be forbidden
	// regardless of the FTC side is zero or not.
	// This is suspicious fraud.
	// We still need to update the IAP side membership based on
	// apple latest transaction.
	// There is a possibility that the FTC side is expired
	// we take it as error because if we allow it,
	// then cheater might be using the same apple id to link
	// to multiple invalid FTC memberships.
	if !iapMember.IsZero() {
		// Here
		// b != 0, a == 0 iap already exists while ftc is zero;
		// b != 0, a != 0 iap already exists and ftc is not zero.
		// It includes both non-zero cases.
		return Membership{}, ErrLinkToMultipleFTC
	}

	// Here b == 0, a != 0.
	// Now the IAP side is zero while the FTC side is not-zero.
	// If the FTC side is no longer valid, it is allowed to have
	// apple_subscription_id set to apple's original transaction memberID.
	// This might erase previous original transaction memberID
	// set on the FTC side. It's ok since it is already invalid.
	if !m.IsValid() {
		return m, nil
	}

	// FTC side is already linked an apple memberID.
	// This might occur when user changed apple memberID and is trying
	// to link to the same FTC account which is linked to old
	// apple memberID.
	if m.IsIAP() {
		return Membership{}, ErrTargetLinkedToOtherIAP
	}

	// FTC side have a valid membership purchased via
	// non-apple channel.
	return Membership{}, ErrHasValidNonIAPMember
}

// FromGiftCard creates a new instance based on a gift card.
func (m Membership) FromGiftCard(c redeem.GiftCard) (Membership, error) {

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

// Snapshot takes a snapshot of membership, usually before modifying it.
func (m Membership) Snapshot(reason enum.SnapshotReason) MemberSnapshot {
	if m.IsZero() {
		return MemberSnapshot{}
	}

	return MemberSnapshot{
		SnapshotID: GenerateSnapshotID(),
		Reason:     reason,
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
