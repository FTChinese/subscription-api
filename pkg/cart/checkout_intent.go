package cart

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// CheckoutIntent decides how user want to purchase a product.
// This is determined by current membership, product and payment method selected.
// If user chooses Ali/Wx, it is a one-time purchase; for stripe it is a subscription.
// `OneTimeKind` and `SubsKind` should not exist at the same time.
type CheckoutIntent struct {
	// What kind of one-time purchase user is trying to create?
	OneTimeKind enum.OrderKind
	// How would user perform a subscription:
	// creating a new one?
	// Just updating it to different billing cycle or tier?
	// Or switching one-time purchase to subscription mode?
	// In the last case, current remaining days should be transferred to add-on
	SubsKind   SubsKind
	PayMethods []enum.PayMethod
}

func NewOneTimeIntent(kind enum.OrderKind) CheckoutIntent {
	return CheckoutIntent{
		OneTimeKind: kind,
		PayMethods: []enum.PayMethod{
			enum.PayMethodAli,
			enum.PayMethodWx,
		},
	}
}

func NewSubsIntent(kind SubsKind) CheckoutIntent {
	return CheckoutIntent{
		SubsKind: kind,
		PayMethods: []enum.PayMethod{
			enum.PayMethodStripe,
		},
	}
}

func (i CheckoutIntent) PermitNewStripe() bool {
	return i.SubsKind == SubsKindNew || i.SubsKind == SubsKindOneTimeToStripe
}

// Contains checks if the payment method contains the specified one.
func (i CheckoutIntent) Contains(m enum.PayMethod) bool {
	for _, v := range i.PayMethods {
		if v == m {
			return true
		}
	}

	return false
}

type CheckoutIntents struct {
	intents []CheckoutIntent
	err     error
}

// Get finds the intent for the specified payment method, or returns error
// if not found.
func (coi CheckoutIntents) Get(m enum.PayMethod) (CheckoutIntent, error) {
	if coi.err != nil {
		return CheckoutIntent{}, coi.err
	}

	if len(coi.intents) == 0 {
		return CheckoutIntent{}, errors.New("cannot determine checkout intent")
	}

	for _, intent := range coi.intents {
		if intent.Contains(m) {
			return intent, nil
		}
	}

	return CheckoutIntent{}, errors.New("cannot determine checkout intent")
}

func NewCheckoutIntents(m reader.Membership, e price.Edition) CheckoutIntents {
	if m.IsExpired() {
		return CheckoutIntents{
			intents: []CheckoutIntent{
				NewOneTimeIntent(enum.OrderKindCreate),
				NewSubsIntent(SubsKindNew),
			},
			err: nil,
		}
	}

	if m.IsInvalidStripe() {
		return CheckoutIntents{
			intents: []CheckoutIntent{
				NewOneTimeIntent(enum.OrderKindCreate),
				NewSubsIntent(SubsKindNew),
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
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindRenew),
					// However, if user want to use Stripe to subscribe,
					// it should be treated as a new subscription,
					// with current remaining subscription time reserved for future use.
					NewSubsIntent(SubsKindOneTimeToStripe),
				},
				err: nil,
			}
		}

		// The product to purchase differs from current one.
		switch e.Tier {
		// Upgrading to premium.
		case enum.TierPremium:
			return CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindUpgrade),
					NewSubsIntent(SubsKindOneTimeToStripe),
				},
				err: nil,
			}

		// Current premium want to buy standard.
		// Only add-on is allowed.
		case enum.TierStandard:
			return CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			}
		}

	case enum.PayMethodStripe:
		// As long as user is subscribed to premium, only add-on is allowed.
		if m.Tier == enum.TierPremium {
			return CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			}
		}

		// User subscribed to standard.
		switch e.Tier {
		// Upgrade to premium could be done via stripe updating.
		case enum.TierPremium:
			return CheckoutIntents{
				intents: []CheckoutIntent{
					NewSubsIntent(SubsKindUpgrade),
				},
				err: nil,
			}

		case enum.TierStandard:
			// For the same edition, no need to update stripe.
			// Only add-on is allowed.
			if m.Cycle == e.Cycle {
				return CheckoutIntents{
					intents: []CheckoutIntent{
						NewOneTimeIntent(enum.OrderKindAddOn),
					},
					err: nil,
				}
			}
			// For same tier, different cycle.
			return CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
					NewSubsIntent(SubsKindSwitchCycle),
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
			intents: []CheckoutIntent{
				NewOneTimeIntent(enum.OrderKindAddOn),
			},
			err: nil,
		}

	case enum.PayMethodB2B:
		return CheckoutIntents{
			intents: []CheckoutIntent{
				NewOneTimeIntent(enum.OrderKindAddOn),
			},
			err: nil,
		}
	}

	return CheckoutIntents{
		intents: nil,
		err:     errors.New("operation not supported"),
	}
}
