package reader

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"net/http"
	"time"

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

func GetTierCode(tier enum.Tier) int64 {
	return tierToCode[tier]
}

// AddOn specifies extra days that can be used after current expiration date reached.
type AddOn struct {
	Standard int64 `json:"standardAddOn" db:"standard_addon"`
	Premium  int64 `json:"premiumAddOn" db:"premium_addon"`
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
	FtcPlanID     null.String    `json:"-" db:"ftc_plan_id"`
	StripeSubsID  null.String    `json:"stripeSubsId" db:"stripe_subs_id"`
	StripePlanID  null.String    `json:"-" db:"stripe_plan_id"`
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

	return m.CompoundID == other.CompoundID &&
		m.StripeSubsID == other.StripeSubsID &&
		m.AppleSubsID.String == other.AppleSubsID.String &&
		m.Tier == other.Tier
}

func (m Membership) IsModified(other Membership) bool {
	if !m.IsEqual(other) {
		return true
	}

	if !m.ExpireDate.Equal(other.ExpireDate.Time) {
		return true
	}

	if m.PaymentMethod != other.PaymentMethod {
		return true
	}

	if m.FtcPlanID != other.FtcPlanID {
		return true
	}

	if m.AutoRenewal != other.AutoRenewal {
		return true
	}

	if m.Status != other.Status {
		return true
	}

	return false
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

// Sync turns legacy vip_type and expire_time into
// member_tier and expire_date columns, or vice versus.
func (m Membership) Sync() Membership {
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

func (m Membership) IsStripe() bool {
	return !m.IsZero() && m.PaymentMethod == enum.PayMethodStripe && m.StripeSubsID.Valid
}

// IsIAP tests whether this membership comes from Apple.
// This is actually not necessary. However, as People are allowed to changed DB directly, if an IAP membership is
// changed to other payment methods, and chances are high that humans only change the payment method column but
// do not nullify the AppleSubsID field, which is probably true since it is hard for human to find out apple's original
// transaction id.
func (m Membership) IsIAP() bool {
	return !m.IsZero() && m.PaymentMethod == enum.PayMethodApple && m.AppleSubsID.Valid
}

func (m Membership) IsB2B() bool {
	return !m.IsZero() && m.PaymentMethod == enum.PayMethodB2B && m.B2BLicenceID.Valid
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
// Kind  |   Membership
// ---------------------------
// Create  | Zero / Status is not active / Expired
// Renewal | Tier === Plan.Tier
// Upgrade | Tier is Standard while Plan.Tier is Premium
func (m Membership) AliWxSubsKind(e product.Edition) (enum.OrderKind, error) {
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
		return enum.OrderKindNull, fmt.Errorf("already subscribed via %s", m.PaymentMethod)
	}

	// Renewal.
	if m.Tier == e.Tier {
		if m.canRenewViaAliWx() {
			return enum.OrderKindRenew, nil
		} else {
			return enum.OrderKindNull, errors.New("beyond max allowed renewal period")
		}
	}

	// Trying to upgrade.
	if m.Tier == enum.TierStandard && e.Tier == enum.TierPremium {
		return enum.OrderKindUpgrade, nil
	}

	// Trying to downgrade.
	// TODO: to allow downgrading, we should establish a system that
	// keep order in reserved state.
	// Only change current membership based on reserved order after expiration date reached.
	// The same approach could be used to handle a valid B2B, IAP or Stripe user creating orders via wx and ali.
	if m.Tier == enum.TierPremium && e.Tier == enum.TierStandard {
		return enum.OrderKindNull, errors.New("downgrading is not supported currently")
	}

	return enum.OrderKindNull, errors.New("cannot determine subscription kind")
}

// StripeSubsKind deduce what kind of subscription a request is trying to create.
// You can create or upgrade a subscription via stripe.
// Cases for permission:
// 1. Membership does not exist;
// 2. Membership exists via alipay or wxpay but expired;
// 3. Membership exits via stripe but is expired and it is not auto renewable.
// 4. A stripe subscription that is not expired, auto renewable, but its current status is not active.
func (m Membership) StripeSubsKind(e product.Edition) (enum.OrderKind, error) {
	// If a membership does not exist, allow create stripe
	// subscription
	if m.IsZero() {
		return enum.OrderKindCreate, nil
	}

	// If current membership already expired.
	if m.IsExpired() {
		return enum.OrderKindCreate, nil
	}

	// If not purchased via stripe.
	if m.PaymentMethod != enum.PayMethodStripe {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "You are already subscribed via non-stripe method",
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		}
	}

	// If user previously subscribed via stripe and canceled, it expiration date is not past yet, and auto renewal is off.
	// In such case we should allow user to create a new subscription of the same tier.
	// If current membership is not expired yet, it could be either due to auto renewal, or expiration date is in the future.
	// Status.IsValid is equal to auto renewal.
	// If it is not auto renewal, it might be in canceled state.
	if !m.Status.IsValid() {
		return enum.OrderKindCreate, nil
	}

	// Auto renewable is not needed.
	if m.Tier == e.Tier {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "You are already subscribed via Stripe",
			Field:   "membership",
			Code:    render.CodeAlreadyExists,
		}
	}

	// current tier != requested tier.
	// If current is premium, requested must be standard.
	if m.Tier == enum.TierPremium {
		return enum.OrderKindNull, &render.ValidationError{
			Message: "Downgrading is not supported currently",
			Field:   "downgrade",
			Code:    render.CodeInvalid,
		}
	}

	// Current is standard, requested must be premium.
	return enum.OrderKindUpgrade, nil
}

// PermitStripeCreate checks whether current membership permit creating subscription via stripe.
// Returned error is either render.ValidationError or render.ResponseError.
func (m Membership) PermitStripeCreate(e product.Edition) error {
	k, err := m.StripeSubsKind(e)
	if err != nil {
		return err
	}

	if k != enum.OrderKindCreate {
		return &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "This endpoint only support creating new subscription",
			Invalid:    nil,
		}
	}

	return nil
}

func (m Membership) PermitStripeUpgrade(e product.Edition) error {
	if m.IsZero() {
		return &render.ResponseError{
			StatusCode: http.StatusNotFound,
			Message:    "The subscription to upgrade not found",
			Invalid:    nil,
		}
	}

	if m.StripeSubsID.IsZero() {
		return &render.ResponseError{
			StatusCode: http.StatusNotFound,
			Message:    "Subscription not created via stripe",
			Invalid:    nil,
		}
	}

	k, err := m.StripeSubsKind(e)
	if err != nil {
		return err
	}

	if k != enum.OrderKindUpgrade {
		return &render.ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "This endpoint only support upgrading an existing stripe subscription",
			Invalid:    nil,
		}
	}

	return nil
}

// Snapshot takes a snapshot of membership, usually before modifying it.
func (m Membership) Snapshot(by Archiver) MemberSnapshot {
	if m.IsZero() {
		return MemberSnapshot{}
	}

	return MemberSnapshot{
		SnapshotID: GenerateSnapshotID(),
		CreatedBy:  null.StringFrom(by.String()),
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
