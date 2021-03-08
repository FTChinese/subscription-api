package reader

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type CheckoutIntents struct {
	intents []cart.CheckoutIntent
	err     error
}

// Get finds the intent for the specified payment method, or returns error
// if not found.
func (coi CheckoutIntents) Get(m enum.PayMethod) (cart.CheckoutIntent, error) {
	if coi.err != nil {
		return cart.CheckoutIntent{}, coi.err
	}

	if len(coi.intents) == 0 {
		return cart.CheckoutIntent{}, fmt.Errorf("illegal checkout via %s", m)
	}

	for _, intent := range coi.intents {
		if intent.Contains(m) {
			return intent, nil
		}
	}

	return cart.CheckoutIntent{}, fmt.Errorf("illegal checkout via %s", m)
}

func NewCheckoutIntents(m Membership, e price.Edition) CheckoutIntents {
	if m.IsExpired() {
		return CheckoutIntents{
			intents: []cart.CheckoutIntent{
				cart.NewOneTimeIntent(enum.OrderKindCreate),
				cart.NewSubsIntent(cart.SubsKindNew),
			},
			err: nil,
		}
	}

	if m.IsInvalidStripe() {
		return CheckoutIntents{
			intents: []cart.CheckoutIntent{
				cart.NewOneTimeIntent(enum.OrderKindCreate),
				cart.NewSubsIntent(cart.SubsKindNew),
			},
			err: nil,
		}
	}

	// Current payment method decides how user could pay for new purchase.
	switch m.PaymentMethod {
	// For membership purchased via Ali/Wx, user could continue to user them.
	case enum.PayMethodAli, enum.PayMethodWx:
		// Renewal if user choosing product of same tier.
		if m.Tier == e.Tier {
			// For one-time purchase, do not allow purchase beyond 3 years.
			if !m.WithinMaxRenewalPeriod() {
				return CheckoutIntents{
					intents: nil,
					err:     errors.New("beyond max allowed renewal period"),
				}
			}

			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindRenew),
					// However, if user want to use Stripe to subscribe,
					// it should be treated as a new subscription,
					// with current remaining subscription time reserved for future use.
					cart.NewSubsIntent(cart.SubsKindOneTimeToStripe),
				},
				err: nil,
			}
		}

		// The product to purchase differs from current one.
		switch e.Tier {
		// Upgrading to premium.
		case enum.TierPremium:
			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindUpgrade),
					cart.NewSubsIntent(cart.SubsKindOneTimeToStripe),
				},
				err: nil,
			}

		// Current premium want to buy standard.
		// Only add-on is allowed.
		case enum.TierStandard:
			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			}
		}

	case enum.PayMethodStripe:
		// As long as user is subscribed to premium, only add-on is allowed.
		if m.Tier == enum.TierPremium {
			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			}
		}

		// User subscribed to standard.
		switch e.Tier {
		// Upgrade to premium could be done via stripe updating.
		case enum.TierPremium:
			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewSubsIntent(cart.SubsKindUpgrade),
				},
				err: nil,
			}

		case enum.TierStandard:
			// For the same edition, no need to update stripe.
			// Only add-on is allowed.
			if m.Cycle == e.Cycle {
				return CheckoutIntents{
					// Not allowed to use stripe since changing the same subscription is meaningless.
					intents: []cart.CheckoutIntent{
						cart.NewOneTimeIntent(enum.OrderKindAddOn),
					},
					err: nil,
				}
			}
			// For same tier, different cycle.
			return CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
					cart.NewSubsIntent(cart.SubsKindSwitchCycle),
				},
				err: nil,
			}
		}

	case enum.PayMethodApple:
		if m.Tier == enum.TierStandard && e.Tier == enum.TierPremium {
			return CheckoutIntents{
				intents: nil,
				err:     errors.New("upgrading apple subscription could only be performed on ios devices"),
			}
		}

		return CheckoutIntents{
			intents: []cart.CheckoutIntent{
				cart.NewOneTimeIntent(enum.OrderKindAddOn),
			},
			err: nil,
		}

	case enum.PayMethodB2B:
		return CheckoutIntents{
			intents: []cart.CheckoutIntent{
				cart.NewOneTimeIntent(enum.OrderKindAddOn),
			},
			err: nil,
		}
	}

	return CheckoutIntents{
		intents: nil,
		err:     errors.New("operation not supported"),
	}
}
