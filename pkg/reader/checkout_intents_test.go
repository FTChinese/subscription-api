package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/price"
	"reflect"
	"testing"
)

func TestNewCheckoutIntents(t *testing.T) {
	type args struct {
		m Membership
		e price.Edition
	}
	tests := []struct {
		name string
		args args
		want CheckoutIntents
	}{
		{
			name: "Expired membership",
			args: args{
				m: Membership{},
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindCreate),
					cart.NewSubsIntent(cart.SubsKindNew),
				},
				err: nil,
			},
		},
		{
			name: "Invalid stripe",
			args: args{
				m: NewMockMemberBuilder("").
					WithSubsStatus(enum.SubsStatusUnpaid).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindCreate),
					cart.NewSubsIntent(cart.SubsKindNew),
				},
				err: nil,
			},
		},
		{
			name: "Onetime renewal",
			args: args{
				m: NewMockMemberBuilder("").Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindRenew),
					cart.NewSubsIntent(cart.SubsKindOneTimeToStripe),
				},
				err: nil,
			},
		},
		{
			name: "Onetime upgrade",
			args: args{
				m: NewMockMemberBuilder("").Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindUpgrade),
					cart.NewSubsIntent(cart.SubsKindOneTimeToStripe),
				},
				err: nil,
			},
		},
		{
			name: "Onetime premium buy addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe can purchase standard addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe can purchase premium addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can upgrade",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewSubsIntent(cart.SubsKindUpgrade),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can purchase addon of same tier",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can switch cycle of different cycle or purchase addon of same tier",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdMonthEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
					cart.NewSubsIntent(cart.SubsKindSwitchCycle),
				},
				err: nil,
			},
		},
		{
			name: "IAP cannot upgrade",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: nil,
				err:     errors.New("upgrading apple subscription could only be performed on ios devices"),
			},
		},
		{
			name: "IAP standard can purchase standard addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "IAP premium can purchase premium addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "IAP premium can purchase standard addon",
			args: args{
				m: NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []cart.CheckoutIntent{
					cart.NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCheckoutIntents(tt.args.m, tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCheckoutIntents() = %v, want %v", got, tt.want)
			}
		})
	}
}
