package paywall

import (
	"github.com/pkg/errors"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Duration contains a membership's expiration time.
// This type exits for compatibility due to expiration time are saved into two columns.
type Duration struct {
	Timestamp  int64
	ExpireDate chrono.Date
}

// NormalizeDate converts unix timestamp to util.Date.
func (d *Duration) NormalizeDate() {
	if d.ExpireDate.IsZero() && d.Timestamp != 0 {
		d.ExpireDate = chrono.DateFrom(time.Unix(d.Timestamp, 0))
	}
}

// CanRenew tests if a membership is allowed to renuew subscription.
// A member could only renew its subscripiton when remaining duration of a membership is shorter than a billing cycle.
// Expire date - now > cycle  --- Renwal is not allowed
// Expire date - now <= cycle --- Can renew
//         now--------------------| Allow
//      |-------- A cycle --------| Expires
// now----------------------------| Deny
func (d Duration) CanRenew(cycle enum.Cycle) bool {
	cycleEnds, err := cycle.TimeAfterACycle(time.Now())

	if err != nil {
		return false
	}

	return d.ExpireDate.Before(cycleEnds)
}

// IsExpired tests if the membership's expiration date is before now.
func (d Duration) IsExpired() bool {
	if d.ExpireDate.IsZero() {
		return true
	}
	// If expire is before now, it is expired.
	return d.ExpireDate.Before(time.Now())
}

// Membership contains a user's membership details
type Membership struct {
	CompoundID string      `json:"-"` // Either FTCUserID or UnionID
	FTCUserID  null.String `json:"-"`
	UnionID    null.String `json:"-"` // For both vip_id_alias and wx_union_id columns.
	Tier       enum.Tier   `json:"tier"`
	Cycle      enum.Cycle  `json:"billingCycle"`
	Duration               // On which date the membership ends
}

// NewMember creates a membership directly for a user.
// This is currently used by activating gift cards.
// If membership is purchased via direct payment channel,
// membership is created from subscription order.
func NewMember(ftcID null.String, unionID null.String) (Membership, error) {
	m := Membership{
		FTCUserID: ftcID,
		UnionID: unionID,
	}

	if ftcID.Valid {
		m.CompoundID = ftcID.String
	} else if unionID.Valid {
		m.CompoundID = unionID.String
	} else {
		return m, errors.New("ftcID and unionID should not both be null")
	}

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
