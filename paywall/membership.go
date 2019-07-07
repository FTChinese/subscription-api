package paywall

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Membership contains a user's membership details
type Membership struct {
	CompoundID string      `json:"-"` // Either FTCUserID or UnionID
	FTCUserID  null.String `json:"-"`
	UnionID    null.String `json:"-"` // For both vip_id_alias and wx_union_id columns.
	Tier       enum.Tier   `json:"tier"`
	Cycle      enum.Cycle  `json:"billingCycle"`
	ExpireDate chrono.Date
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(u UserID) Membership {
	return Membership{
		CompoundID: u.CompoundID,
		FTCUserID:  u.FtcID,
		UnionID:    u.UnionID,
	}
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
	if m.Tier == enum.InvalidTier {
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
