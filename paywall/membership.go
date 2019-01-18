package paywall

import (
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
	UserID   string      `json:"-"`
	UnionID  null.String `json:"-"`
	Tier     enum.Tier   `json:"tier"`
	Cycle    enum.Cycle  `json:"billingCycle"`
	Duration             // On which date the membership ends
}
