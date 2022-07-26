package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type CheckoutIntent struct {
	Kind  SubsIntentKind `json:"kind"`
	Error error          `json:"error"` // A message telling why the Kind is IntentForbidden.
}

var unknownCheckout = CheckoutIntent{
	Kind:  IntentForbidden,
	Error: ErrUnknownPaymentMethod,
}

// NewCheckoutIntentStripe deduces what kind of action
// when user is trying is subscribed via Stripe.
func NewCheckoutIntentStripe(m Membership, item CartItemStripe) CheckoutIntent {
	if m.IsExpired() || m.IsInvalidStripe() {
		return CheckoutIntent{
			Kind:  IntentCreate,
			Error: nil,
		}
	}

	switch m.PaymentMethod {
	// One-off purchase -> Stripe
	case enum.PayMethodAli, enum.PayMethodWx:
		return CheckoutIntent{
			Kind:  IntentOneTimeToAutoRenew,
			Error: nil,
		}

	// Stripe -> Stripe
	case enum.PayMethodStripe:
		// Save tier.
		if m.Tier == item.Recurring.Tier {
			// Save edition
			if m.Cycle == item.Recurring.PeriodCount.EqCycle() {
				if item.HasCoupon() {
					return CheckoutIntent{
						Kind:  IntentApplyCoupon,
						Error: nil,
					}
				}

				return CheckoutIntent{
					Kind:  IntentForbidden,
					Error: ErrAlreadyStripeSubs,
				}
			}
			// Same tier and different cycle.
			return CheckoutIntent{
				Kind:  IntentSwitchInterval,
				Error: nil,
			}
		}

		// Different tier.
		switch item.Recurring.Tier {
		// Current standard to Premium
		case enum.TierPremium:
			//if m.IsTrialing() {
			//	return CheckoutIntent{
			//		Kind:  IntentForbidden,
			//		Error: ErrTrialUpgradeForbidden,
			//	}
			//}
			return CheckoutIntent{
				Kind:  IntentUpgrade,
				Error: nil,
			}
		// Current premium to standard
		case enum.TierStandard:
			return CheckoutIntent{
				Kind:  IntentDowngrade,
				Error: nil,
			}
		}

	case enum.PayMethodApple:
		return CheckoutIntent{
			Kind:  IntentForbidden,
			Error: ErrAlreadyAppleSubs,
		}

	case enum.PayMethodB2B:
		return CheckoutIntent{
			Kind:  IntentForbidden,
			Error: ErrAlreadyB2BSubs,
		}
	}

	return unknownCheckout
}

// NewCheckoutIntentFtc determines what a user can do
// when trying to pay via ali/wx, depending on the current
// membership.
func NewCheckoutIntentFtc(m Membership, p price.FtcPrice) CheckoutIntent {
	if m.IsExpired() || m.IsInvalidStripe() {
		return CheckoutIntent{
			Kind:  IntentCreate,
			Error: nil,
		}
	}

	// What can be done depends on current payment method.
	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		// Renewal if user choosing product of same tier.
		if m.Tier == p.Tier {
			// For one-time purchase, do not allow purchase beyond 3 years.
			if !m.WithinMaxRenewalPeriod() {
				return CheckoutIntent{
					Kind:  IntentForbidden,
					Error: ErrExceedingMaxRenewal,
				}
			}

			return CheckoutIntent{
				Kind:  IntentRenew,
				Error: nil,
			}
		}

		// The product to purchase differs from current one.
		switch p.Tier {
		// Upgrading to premium.
		case enum.TierPremium:
			return CheckoutIntent{
				Kind:  IntentUpgrade,
				Error: nil,
			}

		// Current premium want to buy standard.
		// For Ali/Wx, it is add-on; however, user is allowed to switch to stripe.
		case enum.TierStandard:
			return CheckoutIntent{
				Kind:  IntentAddOn,
				Error: nil,
			}
		}

	case enum.PayMethodStripe:
		// Stripe user purchase same tier of one-time.
		if m.Tier == p.Tier {
			return CheckoutIntent{
				Kind:  IntentAddOn,
				Error: nil,
			}
		}

		switch p.Tier {
		// tripe standard -> onetime premium
		case enum.TierPremium:
			return CheckoutIntent{
				Kind:  IntentForbidden,
				Error: ErrSubsUpgradeViaOneTime,
			}

		case enum.TierStandard:
			// Stripe premium -> onetime standard
			return CheckoutIntent{
				Kind:  IntentAddOn,
				Error: nil,
			}
		}

	case enum.PayMethodApple:
		if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
			return CheckoutIntent{
				Kind:  IntentForbidden,
				Error: ErrSubsUpgradeViaOneTime,
			}
		}

		return CheckoutIntent{
			Kind:  IntentAddOn,
			Error: nil,
		}

	case enum.PayMethodB2B:
		if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
			return CheckoutIntent{
				Kind:  IntentForbidden,
				Error: ErrB2BUpgradeViaOneTime,
			}
		}

		return CheckoutIntent{
			Kind:  IntentRenew,
			Error: nil,
		}
	}

	return unknownCheckout
}

func NewCheckoutIntentApple(m Membership) CheckoutIntent {
	if m.IsExpired() || m.IsInvalidStripe() {
		return CheckoutIntent{
			Kind:  IntentCreate,
			Error: nil,
		}
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		return CheckoutIntent{
			Kind:  IntentOneTimeToAutoRenew,
			Error: nil,
		}

	case enum.PayMethodStripe:
		return CheckoutIntent{
			Kind:  IntentForbidden,
			Error: errors.New("iap is not allowed to override a valid stripe subscription"),
		}

	case enum.PayMethodApple:
		return CheckoutIntent{
			Kind:  IntentRenew,
			Error: nil,
		}
	}

	return CheckoutIntent{
		Kind:  IntentOneTimeToAutoRenew,
		Error: nil,
	}
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p CheckoutIntent) Value() (driver.Value, error) {

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (p *CheckoutIntent) Scan(src interface{}) error {
	if src == nil {
		*p = CheckoutIntent{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp CheckoutIntent
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to CheckoutIntent")
	}
}
