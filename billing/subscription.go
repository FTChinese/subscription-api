package billing

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

type Subscription struct {
	// Fields common to all.
	OrderID string `json:"id"`
	UserID
	Charge

	Plan    Plan    `json:"plan"`
	Coupon  Coupon  `json:"coupon"`
	Upgrade Upgrade `json:"upgrade,omitempty"`

	TierToBuy    enum.Tier  `json:"tier"`  // Deprecate
	BillingCycle enum.Cycle `json:"cycle"` // Deprecate

	CycleCount int64 `json:"cycleCount"` // Default to 1. Change it for upgrade
	ExtraDays  int64 `json:"extraDays"`  // Deprecate. Default to 1. Change it for upgraded.
	TrialDays  int64 `json:"trialDays"`

	// The category of this order.
	Kind SubsKind `json:"usageType"`

	// Fields only applicable to upgrade.
	UpgradeSource  []string   `json:"-"`       // Deprecate. for upgrade
	UpgradeBalance null.Float `json:"balance"` // Deprecate. for upgrade

	// Payment method
	PaymentMethod enum.PayMethod `json:"paymentMethod"`
	WxAppID       null.String    `json:"-"` // Wechat specific

	CreatedAt chrono.Time `json:"createdAt"`

	// Fields populated only after payment finished.
	ConfirmedAt chrono.Time `json:"confirmedAt"` // When the payment is confirmed.
	StartDate   chrono.Date `json:"startDate"`   // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate     chrono.Date `json:"endDate"`     // Membership end date for this order. Depends on start date.
}
