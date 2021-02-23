package reader

import (
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"math"
	"time"

	"github.com/FTChinese/subscription-api/pkg/price"

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
	price.Edition
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
	addon.ReservedDays
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

// RemainingDays calculates how many day left up until now.
// If the returned days is less than 0, the membership is expired
// if it is not auto renewable.
func (m Membership) RemainingDays() int64 {
	h := time.Until(m.ExpireDate.Time).Hours()

	return int64(math.Ceil(h / 24))
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
	// PROBLEM: if both sides have data but they are not synced, the discrepancy will be carried forward.
	return m
}

// IsOneTime checks whether current membership is purchased
// via alipay or wechat pay.
func (m Membership) IsOneTime() bool {
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

func (m Membership) IsInvalidStripe() bool {
	return m.IsStripe() && (m.Status == enum.SubsStatusIncompleteExpired || m.Status == enum.SubsStatusPastDue || m.Status == enum.SubsStatusCanceled || m.Status == enum.SubsStatusUnpaid)
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

// WithinMaxRenewalPeriod test if current membership is allowed to renew for wxpay or alipay.
// now <= expire date <= 3 years later
func (m Membership) WithinMaxRenewalPeriod() bool {
	today := time.Now().Truncate(24 * time.Hour)
	threeYearsLater := today.AddDate(3, 0, 0)

	// It should include today and the date three year later.
	return !m.ExpireDate.Before(today) && !m.ExpireDate.After(threeYearsLater)
}

// Snapshot takes a snapshot of membership, usually before modifying it.
func (m Membership) Snapshot(by Archiver) MemberSnapshot {
	if m.IsZero() {
		return MemberSnapshot{}
	}

	return MemberSnapshot{
		SnapshotID: db.SnapshotID(),
		CreatedBy:  null.StringFrom(by.String()),
		CreatedUTC: chrono.TimeNow(),
		Membership: m,
	}
}
