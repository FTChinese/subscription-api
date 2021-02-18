package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type CheckoutIntent struct {
	OrderKind  enum.OrderKind
	PayMethods []enum.PayMethod
}

type CheckoutIntents []CheckoutIntent

func (i CheckoutIntents) FindIntent(m enum.PayMethod) (CheckoutIntent, error) {
	switch len(i) {
	case 0:
		return CheckoutIntent{}, errors.New("cannot determine checkout intent")

	case 1:
		return i[0], nil
	}

	for _, intent := range i {
		for _, pm := range intent.PayMethods {
			if pm == m {
				return intent, nil
			}
		}
	}

	return CheckoutIntent{}, errors.New("cannot determine checkout intent")
}

func NewCheckoutIntents(m reader.Membership, e price.Edition) (CheckoutIntents, error) {
	if m.IsExpired() {
		return []CheckoutIntent{
			{
				OrderKind: enum.OrderKindCreate,
				PayMethods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
					enum.PayMethodStripe,
				},
			},
		}, nil
	}

	if m.IsInvalidStripe() {
		return []CheckoutIntent{
			{
				OrderKind: enum.OrderKindCreate,
				PayMethods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
					enum.PayMethodStripe,
				},
			},
		}, nil
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		// Renewal
		if m.Tier == e.Tier {
			if !m.WithinMaxRenewalPeriod() {
				return nil, errors.New("beyond max allowed renewal period")
			}

			return []CheckoutIntent{
				{
					OrderKind: enum.OrderKindRenew,
					PayMethods: []enum.PayMethod{
						enum.PayMethodAli,
						enum.PayMethodWx,
						enum.PayMethodStripe,
					},
				},
			}, nil
		}

		switch e.Tier {
		case enum.TierPremium:
			return []CheckoutIntent{
				{
					OrderKind: enum.OrderKindUpgrade,
					PayMethods: []enum.PayMethod{
						enum.PayMethodAli,
						enum.PayMethodWx,
						enum.PayMethodStripe,
					},
				},
			}, nil

		case enum.TierStandard:
			return []CheckoutIntent{
				{
					OrderKind: enum.OrderKindAddOn,
					PayMethods: []enum.PayMethod{
						enum.PayMethodAli,
						enum.PayMethodWx,
					},
				},
			}, nil
		}

	case enum.PayMethodStripe:
		// As long as user is subscribed to premium, only add-on is allowed.
		if m.Tier == enum.TierPremium {
			return []CheckoutIntent{
				{
					OrderKind: enum.OrderKindAddOn,
					PayMethods: []enum.PayMethod{
						enum.PayMethodAli,
						enum.PayMethodWx,
					},
				},
			}, nil
		}

		// User subscribed to standard.
		switch e.Tier {
		// Upgrade to premium could be done via stripe updating.
		case enum.TierPremium:
			return []CheckoutIntent{
				{
					OrderKind:  enum.OrderKindUpgrade,
					PayMethods: []enum.PayMethod{enum.PayMethodStripe},
				},
			}, nil

		case enum.TierStandard:
			// For the same edition, no need to update stripe.
			// Only add-on is allowed.
			if m.Cycle == e.Cycle {
				return []CheckoutIntent{
					{
						OrderKind: enum.OrderKindAddOn,
						PayMethods: []enum.PayMethod{
							enum.PayMethodAli,
							enum.PayMethodWx,
						},
					},
				}, nil
			}
			// For same tier, different cycle.
			return []CheckoutIntent{
				{
					OrderKind: enum.OrderKindAddOn,
					PayMethods: []enum.PayMethod{
						enum.PayMethodAli,
						enum.PayMethodWx,
					},
				},
				{
					OrderKind: enum.OrderKindSwitchCycle,
					PayMethods: []enum.PayMethod{
						enum.PayMethodStripe,
					},
				},
			}, nil
		}

	case enum.PayMethodApple:
		if m.Tier == enum.TierStandard && e.Tier == enum.TierPremium {
			return nil, errors.New("upgrading apple subscription could only be performed on ios devices")
		}

		return []CheckoutIntent{
			{
				OrderKind: enum.OrderKindAddOn,
				PayMethods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
				},
			},
		}, nil

	case enum.PayMethodB2B:
		return []CheckoutIntent{
			{
				OrderKind: enum.OrderKindAddOn,
				PayMethods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
				},
			},
		}, nil
	}

	return nil, errors.New("operation not supported")
}
