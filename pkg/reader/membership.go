package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"log"
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
	ids.UserIDs
	price.Edition
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
	addon.AddOn
	VIP bool `json:"vip" db:"is_vip"`
}

// NewMembership attaches membership directly to a user,
// without any means of payment.
func NewMembership(ba account.BaseAccount, params input.MemberParams) Membership {
	return Membership{
		UserIDs: ba.CompoundIDs(),
		Edition: price.Edition{
			Tier:  params.Tier,
			Cycle: params.Cycle,
		},
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    params.ExpireDate,
		PaymentMethod: params.PayMethod,
		FtcPlanID:     null.StringFrom(params.PriceID),
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		AddOn:         addon.AddOn{},
		VIP:           false,
	}.Sync()
}

func (m Membership) Update(params input.MemberParams) Membership {
	m.Tier = params.Tier
	m.Cycle = params.Cycle
	m.ExpireDate = params.ExpireDate
	m.PaymentMethod = params.PayMethod

	return m
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

// IsLinked checks whether this membership has both FTC id
// and wechat id.
// Do not call this on a merged result since the merging does not link ids.
func (m Membership) IsLinked() bool {
	return m.FtcID.Valid && m.UnionID.Valid && m.FtcID.String != m.UnionID.String
}

// IsFtcOnly checks whether the membership is purchased via
// FTC account only.
// Stripe, Apple and B2B memberships are only allowed to be
// purchased via FTC account.
// When performing unlinking, you should never give such
// kind of membership to the wechat side.
func (m Membership) IsFtcOnly() bool {
	return m.PaymentMethod == enum.PayMethodStripe || m.PaymentMethod == enum.PayMethodB2B || m.PaymentMethod == enum.PayMethodApple
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

func (m Membership) IsTrialing() bool {
	return m.Status == enum.SubsStatusTrialing
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

// OrderKindOfOneTime deduces what kind of order user is trying to create when paying via Ali/Wx.
func (m Membership) OrderKindOfOneTime(e price.Edition) (enum.OrderKind, error) {
	if m.IsExpired() || m.IsInvalidStripe() {
		return enum.OrderKindCreate, nil
	}

	// What can be done depends on current payment method.
	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		// Renewal if user choosing product of same tier.
		if m.Tier == e.Tier {
			// For one-time purchase, do not allow purchase beyond 3 years.
			if !m.WithinMaxRenewalPeriod() {
				return enum.OrderKindNull, errors.New("exceeding allowed max renewal period")
			}

			return enum.OrderKindRenew, nil
		}

		// The product to purchase differs from current one.
		switch e.Tier {
		// Upgrading to premium.
		case enum.TierPremium:
			return enum.OrderKindUpgrade, nil

		// Current premium want to buy standard.
		// For Ali/Wx, it is add-on; however, user is allowed to switch to stripe.
		case enum.TierStandard:
			return enum.OrderKindAddOn, nil
		}
	case enum.PayMethodStripe, enum.PayMethodApple:
		if m.Tier == enum.TierStandard && e.Tier == enum.TierPremium {
			return enum.OrderKindNull, errors.New("subscription mode cannot use one-time purchase to upgrade")
		}
		return enum.OrderKindAddOn, nil
	case enum.PayMethodB2B:
		return enum.OrderKindAddOn, nil
	}

	return enum.OrderKindNull, errors.New("unknown current membership status")
}

// OfferKindsEnjoyed guesses all candidate discounts a user could enjoy.
// This is used as a filter to select an applicable offer.
func (m Membership) OfferKindsEnjoyed() []price.OfferKind {
	if m.IsZero() {
		return []price.OfferKind{
			price.OfferKindPromotion,
			price.OfferKindIntroductory,
		}
	}

	// If current membership is expired, user could enjoy
	// promotion offer (which might not exist), and win back offer
	if m.IsExpired() {
		return []price.OfferKind{
			price.OfferKindPromotion,
			price.OfferKindWinBack,
		}
	}

	// If current membership is not expired, user could enjoy
	// promotion offer (which might not exist), and retention offer
	return []price.OfferKind{
		price.OfferKindPromotion,
		price.OfferKindRetention,
	}
}

func (m Membership) EnjoyOffer(o price.Discount) bool {
	// Empty offer is valid.
	if o.IsZero() {
		return true
	}

	for _, v := range m.OfferKindsEnjoyed() {
		if v == o.Kind {
			return true
		}
	}

	return false
}

// SubsKindOfStripe deduces what kind of subscription user is trying ot create when paying via Stripe.
func (m Membership) SubsKindOfStripe(e price.Edition) (SubsKind, error) {
	if e.Tier == enum.TierNull || e.Cycle == enum.CycleNull {
		return SubsKindZero, errors.New("unknown edition of stripe price")
	}

	if m.IsExpired() || m.IsInvalidStripe() {
		return SubsKindNew, nil
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		return SubsKindOneTimeToSub, nil

	case enum.PayMethodStripe:
		// If already a premium, can do nothing.
		if m.Tier == enum.TierPremium {
			return SubsKindZero, errors.New("already subscribed via stripe")
		}
		// Not premium, then must be standard.
		switch e.Tier {
		// Standard upgrade to premium
		case enum.TierPremium:
			if m.IsTrialing() {
				return SubsKindZero, errors.New("upgrading in trialing period is not allowed")
			}
			return SubsKindUpgrade, nil

		// Standard to standard
		case enum.TierStandard:
			if m.Cycle == e.Cycle {
				return SubsKindZero, errors.New("already subscribed via stripe")
			}

			// Standard changing billing cycle.
			return SubsKindSwitchCycle, nil
		}

	case enum.PayMethodApple:
		return SubsKindZero, errors.New("already subscribed via apple")

	case enum.PayMethodB2B:
		return SubsKindZero, errors.New("already subscribed via stripe")
	}

	return SubsKindZero, errors.New("unknown payment for current subscription")
}

// SubsKindByApple deduces how to handle user's current membership if one exists when Apple webhook arrives.
func (m Membership) SubsKindByApple() (SubsKind, error) {
	if m.IsExpired() || m.IsInvalidStripe() {
		return SubsKindNew, nil
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		return SubsKindOneTimeToSub, nil

	case enum.PayMethodStripe:
		return SubsKindZero, errors.New("iap is not allowed to override a valid stripe subscription")

	case enum.PayMethodApple:
		return SubsKindRefresh, nil
	}

	return SubsKindOneTimeToSub, nil
}

func (m Membership) WithInvoice(userID ids.UserIDs, inv invoice.Invoice) (Membership, error) {
	if inv.IsZero() {
		return m, nil
	}

	// For add-on invoice, only update the add-on part
	// while keep the rest as is since current membership
	// might comes from Stripe or Apple.
	// For upgrading's carry over, we also handle it this way.
	if inv.OrderKind == enum.OrderKindAddOn {
		return m.PlusAddOn(addon.New(inv.Tier, inv.TotalDays())), nil
	}

	// The invoice should have been consumed utc set before updating membership.
	if !inv.IsConsumed() {
		return Membership{}, errors.New("invoice not finalized")
	}

	// If the invoice is not intended for add-on, it must have period set.
	return Membership{
		UserIDs:       userID,
		Edition:       inv.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(inv.EndUTC.Time),
		PaymentMethod: inv.PaymentMethod,
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		AddOn:         m.AddOn, // For upgrade, the carried over part is not added.
	}.Sync(), nil
}

// Merge merges two memberships.
// The returned data is incomplete since if one membership
// is zero, we cannot know user's ftc id or union id.
// Only the merged Account knows all ids.
// Case matrix:
// --------------------------------------------
// FTC\WX      | None  | Not-Expired  | Expired
// --------------------------------------------
// None        |  Y    |    Y         |  Y
// Not-Expired |  Y    |    N         |  Y
// Expired     |  Y    |    Y         |  Y
// --------------------------------------------
// From the above figure we can see the focus is whether
// an account has membership and whether it is expired or
// not.
// A Stripe or Apple membership always falls into the
// Not-Expired group.
func (m Membership) Merge(other Membership) (Membership, error) {
	// This includes both zero values.
	// If the two memberships are the same one, noop.
	if m.IsEqual(other) {
		return m, nil
	}

	// At least one of the membership is not zero value.
	if m.IsLinked() || other.IsLinked() {
		return Membership{}, &render.ValidationError{
			Message: "at least one of the account's membership is linked to a 3rd account",
			Field:   "membership_link",
			Code:    render.CodeAlreadyExists,
		}
	}

	// Now both members are not linked.
	// If both are not expired, deny merging.
	if !m.IsExpired() && !other.IsExpired() {
		log.Print("Error: Both not expired")

		return Membership{}, &render.ValidationError{
			Message: "accounts with valid memberships cannot be linked",
			Field:   "membership_both_valid",
			Code:    render.CodeAlreadyExists,
		}
	}

	// Now both members are not linked.
	// Then cannot be both empty.
	// At least one exists and not expired.
	if m.ExpireDate.Before(other.ExpireDate.Time) {
		return other.PlusAddOn(m.AddOn), nil
	} else {
		return m.PlusAddOn(other.AddOn), nil
	}
}
