package pw

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type CheckoutIntent struct {
	Kind  reader.SubsKind `json:"kind"`
	Error error           `json:"error"`
}

var unknownCheckout = CheckoutIntent{
	Kind:  reader.SubsKindForbidden,
	Error: reader.ErrUnknownPaymentMethod,
}

// NewCheckoutIntentStripe deduces what kind of action
// when user is trying is subscribed via Stripe.
func NewCheckoutIntentStripe(m reader.Membership, p price.StripePrice) CheckoutIntent {
	if m.IsExpired() || m.IsInvalidStripe() {
		return CheckoutIntent{
			Kind:  reader.SubsKindCreate,
			Error: nil,
		}
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		return CheckoutIntent{
			Kind:  reader.SubsKindOneTimeToAutoRenew,
			Error: nil,
		}

	case enum.PayMethodStripe:
		// Save tier.
		if p.Tier == m.Tier {
			// Save edition
			if p.PeriodCount.EqCycle() == m.Cycle {
				return CheckoutIntent{
					Kind:  reader.SubsKindForbidden,
					Error: reader.ErrAlreadyStripeSubs,
				}
			}
			// Same tier and different cycle.
			return CheckoutIntent{
				Kind:  reader.SubsKindSwitchInterval,
				Error: nil,
			}
		}

		// Different tier.
		switch p.Tier {
		// Current standard to Premium
		case enum.TierPremium:
			if m.IsTrialing() {
				return CheckoutIntent{
					Kind:  reader.SubsKindForbidden,
					Error: reader.ErrTrialUpgradeForbidden,
				}
			}
			return CheckoutIntent{
				Kind:  reader.SubsKindUpgrade,
				Error: nil,
			}
		// Current premium to standard
		case enum.TierStandard:
			return CheckoutIntent{
				Kind:  reader.SubsKindDowngrade,
				Error: nil,
			}
		}

	case enum.PayMethodApple:
		return CheckoutIntent{
			Kind:  reader.SubsKindForbidden,
			Error: reader.ErrAlreadyAppleSubs,
		}

	case enum.PayMethodB2B:
		return CheckoutIntent{
			Kind:  reader.SubsKindForbidden,
			Error: reader.ErrAlreadyB2BSubs,
		}
	}

	return unknownCheckout
}

func NewCheckoutIntentFtc(m reader.Membership, p price.FtcPrice) CheckoutIntent {
	if m.IsExpired() || m.IsInvalidStripe() {
		return CheckoutIntent{
			Kind:  reader.SubsKindCreate,
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
					Kind:  reader.SubsKindForbidden,
					Error: errors.New("exceeding allowed max renewal period"),
				}
			}

			return CheckoutIntent{
				Kind:  reader.SubsKindRenew,
				Error: nil,
			}
		}

		// The product to purchase differs from current one.
		switch p.Tier {
		// Upgrading to premium.
		case enum.TierPremium:
			return CheckoutIntent{
				Kind:  reader.SubsKindUpgrade,
				Error: nil,
			}

		// Current premium want to buy standard.
		// For Ali/Wx, it is add-on; however, user is allowed to switch to stripe.
		case enum.TierStandard:
			return CheckoutIntent{
				Kind:  reader.SubsKindAddOn,
				Error: nil,
			}
		}

	case enum.PayMethodStripe:
		// Stripe user purchase same tier of one-time.
		if m.Tier == p.Tier {
			return CheckoutIntent{
				Kind:  reader.SubsKindAddOn,
				Error: nil,
			}
		}

		switch p.Tier {
		// tripe standard -> onetime premium
		case enum.TierPremium:
			return CheckoutIntent{
				Kind:  reader.SubsKindForbidden,
				Error: errors.New("subscription mode cannot use one-time purchase to upgrade"),
			}

		case enum.TierStandard:
			// Stripe premium -> onetime standard
			return CheckoutIntent{
				Kind:  reader.SubsKindAddOn,
				Error: nil,
			}
		}

	case enum.PayMethodApple:
		if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
			return CheckoutIntent{
				Kind:  reader.SubsKindForbidden,
				Error: errors.New("subscription mode cannot use one-time purchase to upgrade"),
			}
		}

		return CheckoutIntent{
			Kind:  reader.SubsKindAddOn,
			Error: nil,
		}

	case enum.PayMethodB2B:
		if m.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
			return CheckoutIntent{
				Kind:  reader.SubsKindForbidden,
				Error: errors.New("corporate subscription cannot use retail payment to upgrade"),
			}
		}

		return CheckoutIntent{
			Kind:  reader.SubsKindRenew,
			Error: nil,
		}
	}

	return unknownCheckout
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
