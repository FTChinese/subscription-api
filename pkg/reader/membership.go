package reader

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"time"

	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/product"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
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
// A membership might be create from various sources:
// * Alipay / Wxpay - Classified under FTC retail
// * B2B
// * Stripe
// * Apple IAP
// We should keep those sources mutually exclusive.
type Membership struct {
	MemberID
	product.Edition
	LegacyTier    null.Int       `json:"-" db:"vip_type"`
	LegacyExpire  null.Int       `json:"-" db:"expire_time"`
	ExpireDate    chrono.Date    `json:"expireDate" db:"expire_date"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	FtcPlanID     null.String    `json:"ftcPlanId" db:"ftc_plan_id"`
	StripeSubsID  null.String    `json:"stripeSubsId" db:"stripe_subs_id"`
	StripePlanID  null.String    `json:"stripePlanId" db:"stripe_plan_id"`
	AutoRenewal   bool           `json:"autoRenew" db:"auto_renewal"`
	// This is used to save stripe subscription status.
	// Since wechat and alipay treats everything as one-time purchase, they do not have a complex state machine.
	// If we could integrate apple in-app purchase, this column
	// might be extended to apple users.
	// Only `active` should be treated as valid member.
	// Wechat and alipay defaults to `active` for backward compatibility.
	Status       enum.SubsStatus `json:"status" db:"subs_status"`
	AppleSubsID  null.String     `json:"appleSubsId" db:"apple_subs_id"`
	B2BLicenceID null.String     `json:"b2bLicenceId" db:"b2b_licence_id"`
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
	return m.ExpireDate.Before(time.Now().Truncate(24*time.Hour)) && !m.AutoRenewal
}

func (m Membership) IsEqual(other Membership) bool {
	if m.IsZero() && other.IsZero() {
		return true
	}

	return m.CompoundID == other.CompoundID && m.StripeSubsID == other.StripeSubsID && m.AppleSubsID.String == other.AppleSubsID.String && m.Tier == other.Tier
}

// isLegacyOnly checks whether the edition information only comes from
// LegacyTier and LegacyExpire fields.
func (m Membership) isLegacyOnly() bool {
	if m.LegacyExpire.Valid && m.LegacyTier.Valid && m.ExpireDate.IsZero() && m.Tier == enum.TierNull {
		return true
	}

	return false
}

// isAPIOnly checks whether the edition information only comes from
// Tier and Cycle fields.
func (m Membership) isAPIOnly() bool {
	if (m.LegacyExpire.IsZero() && m.LegacyTier.IsZero()) && (!m.ExpireDate.IsZero() && m.Tier != enum.TierNull) {
		return true
	}

	return false
}

// Normalize turns legacy vip_type and expire_time into
// member_tier and expire_date columns, or vice versus.
func (m Membership) Normalize() Membership {
	if m.IsZero() {
		return m
	}

	// Syn from legacy format to api created columns
	if m.isLegacyOnly() {
		// Note the conversion is not exactly the same moment since Golang converts Unix in local time.
		expireDate := time.Unix(m.LegacyExpire.Int64, 0)

		m.ExpireDate = chrono.DateFrom(expireDate)
		m.Tier = codeToTier[m.LegacyTier.Int64]
		// m.Cycle cannot be determined

		return m
	}

	// Sync from api columns to legacy column
	if m.isAPIOnly() {
		m.LegacyExpire = null.IntFrom(m.ExpireDate.Unix())
		m.LegacyTier = null.IntFrom(tierToCode[m.Tier])

		return m
	}

	// Otherwise do not touch it.
	return m
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
	return m.AppleSubsID.Valid
}

func (m Membership) IsB2B() bool {
	return m.B2BLicenceID.Valid
}

func (m Membership) IsValidPremium() bool {
	return m.Tier == enum.TierPremium && !m.IsExpired()
}

// canRenewViaAliWx test if current membership is allowed to renew for wxpay or alipay.
// now <= expire date <= 3 years later
func (m Membership) canRenewViaAliWx() bool {
	// If m does not exist, or not create via alipay or wxpay.
	if m.IsZero() || !m.IsAliOrWxPay() {
		return false
	}

	today := time.Now().Truncate(24 * time.Hour)
	threeYearsLater := today.AddDate(3, 0, 0)

	// It should include today and the date three year later.
	return !m.ExpireDate.Before(today) && !m.ExpireDate.After(threeYearsLater)
}

// AliWxSubsKind determines what kind of order a user is creating based on existing membership.
// SubsKind  |   Membership
// ---------------------------
// Create  | Zero / Status is not active / Expired
// Renewal | Tier === Plan.Tier
// Upgrade | Tier is Standard while Plan.Tier is Premium
// TODO: should we deny B2B switching to retailing?
func (m Membership) AliWxSubsKind(e product.Edition) (enum.OrderKind, *render.ValidationError) {
	if m.IsZero() {
		return enum.OrderKindCreate, nil
	}

	// If it is purchased via stripe but the status is not valid.
	if m.PaymentMethod == enum.PayMethodStripe && !m.Status.IsValid() {
		return enum.OrderKindCreate, nil
	}

	// If an existing member expired, treat it as a new member.
	if m.IsExpired() {
		return enum.OrderKindCreate, nil
	}

	// Member is not expired. We need to consider current payment method.
	// If current membership is purchased via methods other than wx or ali, requesting to pay via wx or ali is invalid.
	if !m.IsAliOrWxPay() {
		return enum.OrderKindNull, &render.ValidationError{
			Message: fmt.Sprintf("Already subscribed via %s", m.PaymentMethod),
			Field:   "paymentMethod",
			Code:    render.CodeInvalid,
		}
	}

	// Renewal.
	if m.Tier == e.Tier {
		if m.canRenewViaAliWx() {
			return enum.OrderKindRenew, nil
		} else {
			return enum.OrderKindNull, &render.ValidationError{
				Message: "Already have a very long membership duration",
				Field:   "renewal",
				Code:    render.CodeInvalid,
			}
		}
	}

	// Trying to upgrade.
	if m.Tier == enum.TierStandard && e.Tier == enum.TierPremium {
		return enum.OrderKindUpgrade, nil
	}

	// Trying to downgrade.
	if m.Tier == enum.TierPremium && e.Tier == enum.TierStandard {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "Downgrading is forbidden.",
			Field:   "downgrade",
			Code:    render.CodeInvalid,
		}
	}

	return enum.OrderKindNull, &render.ValidationError{
		Message: "Cannot determine subscription kind.",
		Field:   "order",
		Code:    render.CodeInvalid,
	}
}

// ================================
// The following section handles stripe.
// ================================

// StripeSubsKind deduce what kind of subscription a request is trying to create.
// You can create or upgrade a subscripiton via stripe.
// Cases for permission:
// 1. Membership does not exist;
// 2. Membership exists via alipay or wxpay but expired;
// 3. Membership exits via stripe but is expired and it is not auto renewable.
// 4. A stripe subscription that is not expired, auto renewable, but its current status is not active.
func (m Membership) StripeSubsKind(e product.Edition) (enum.OrderKind, *render.ValidationError) {
	// If a membership does not exist, allow create stripe
	// subscription
	if m.IsZero() {
		return enum.OrderKindCreate, nil
	}

	if m.IsExpired() {
		return enum.OrderKindCreate, nil
	}

	// If not purchased via stripe.
	if m.PaymentMethod != enum.PayMethodStripe {
		return enum.OrderKindNull, &render.ValidationError{
			Message: fmt.Sprintf("Already subscribed via %s", m.PaymentMethod),
			Field:   "paymentMethod",
			Code:    render.CodeInvalid,
		}
	}

	// Now it is purchase via stripe.
	// Check subscription status.
	if !m.Status.IsValid() {
		return enum.OrderKindCreate, nil
	}

	// Stripe subscription is still valid
	if m.Tier == e.Tier {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "Already subscribed via Stripe",
			Field:   "tier",
			Code:    render.CodeAlreadyExists,
		}
	}

	// Trying to switch tier.
	if m.Tier == enum.TierPremium {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "Downgrading is forbidden.",
			Field:   "downgrade",
			Code:    render.CodeInvalid,
		}
	}

	return enum.OrderKindUpgrade, nil
}

// =======================================
// The following section handles Apple IAP
// =======================================

// ValidateMergeIAP checks if it is allowed to merge iap membership into an FTC membership.
// Only two cases are allowed to merge:
// * Both sides refer to the same membership (including zero value);
// * IAP side is zero and FTC side non-zero but invalid.
// As long as link is allowed to proceed, two side cannot both have memberships simultaneously.
// We only need to take a snapshot of ftc side if it exists.
//
// | FTC\IAP     | None   | Not-Expired | Expired |
// | ----------- | ------ | ----------- | --------|
// | None        |  Y     |      N      |  N      |
// | Not-Expired |  N     |      N      |  N      |
// | Expired     |  Y     |      N      |  N      |
//
// Row 2 Column 2 has an exception:
// If payMethod is null, ftc side expire time is not after iap side, it is probably comes from IAP.
func (m Membership) ValidateMergeIAP(iapMember Membership, s apple.Subscription) error {
	// Equal means either both are zero values, or they refer to the same instance.
	// In such case it is fine to return any of them.
	// The caller should then check whether the returned value is zero.
	// For zero value, build a new membership based on IAP transaction.
	// If any of them is not zero, it indicates they already linked.
	// We should stop processing.
	if iapMember.IsEqual(m) {
		if m.IsZero() {
			return nil
		}

		// Tell calle to stop processing.
		return ErrIAPFtcLinked
	}

	// If the two sides are not equal, they must be totally different memberships and the two sides cannot be both empty.
	//
	// a != b might indicates those cases:
	// * a == 0, b != 0;
	// * a != 0, b != 0;
	// * a != 0, b == 0.
	//
	// Case 1 and 2:
	// The presence of IAP side itself indicates it is already linked to a FTC account.
	// Now it is trying to link to another FTC account since the two sides are not the same one.
	// Such action should be denied regardless of whether the FTC side is valid or not, and it is mostly a fraudulent behavior.
	// If an exiting linked IAP is trying to switch the linked FTC account, it falls into this category and user should first perform unlink.
	// In such case we still need to update the IAP side membership based on apple latest transaction.
	if !iapMember.IsZero() {
		return &render.ValidationError{
			Message: "An apple subscription cannot link to multiple FTC accounts",
			Field:   "iap_membership",
			Code:    "already_linked",
		}
	}

	// Case 3:
	// Current IAP side is empty, then FTC side must not be empty.
	// We need to consider 2 cases here:
	// * FTC side is created via another IAP. In such case we should deny it and user should manually unlink that IAP before linking to this one.
	if m.IsIAP() {
		return &render.ValidationError{
			Message: "Target ftc account is already linked to another apple subscription",
			Field:   "ftc_membership",
			Code:    "already_linked",
		}
	}

	// * FTC side is non-IAP.
	// Then check whether it is expired.
	// If the FTC side is still valid, merging is not allowed since it will override valid data.
	if !m.IsExpired() {
		// An edge case here: if the data is in legacy format and payMethod is null, which might be created by wxpay or
		// or alipay, or might be manually created by customer service, we could not determine whether the linked should
		// be allowed or not.
		// In such case, we will compare the expiration date.
		// If apple's expiration date comes later, allow the FTC
		// side to be overridden; otherwise we shall keep the FTC
		// side intact.
		if m.PaymentMethod == enum.PayMethodNull {
			if m.ExpireDate.Before(s.ExpiresDateUTC.Time) {
				return nil
			}
		}

		return &render.ValidationError{
			Message: "Target ftc account already has a valid non-iap membership",
			Field:   "ftc_membership",
			Code:    "valid_non_iap",
		}
	}

	// Otherwise merging to an exiting expired non-iap FTC membership is allowed.
	return nil
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
