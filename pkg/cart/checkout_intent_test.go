package cart

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"reflect"
	"testing"
)

func Test_formatMethods(t *testing.T) {
	type args struct {
		methods []enum.PayMethod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Stringify an array of one payment methods",
			args: args{
				methods: []enum.PayMethod{
					enum.PayMethodAli,
				},
			},
			want: "alipay",
		},
		{
			name: "Stringify an array of payment methods",
			args: args{
				methods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
					enum.PayMethodStripe,
				},
			},
			want: "alipay, wechat or stripe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatMethods(tt.args.methods); got != tt.want {
				t.Errorf("formatMethods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckoutIntents_Get(t *testing.T) {
	type fields struct {
		intents []CheckoutIntent
		err     error
	}
	type args struct {
		m enum.PayMethod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CheckoutIntent
		wantErr bool
	}{
		{
			name: "Find intent by payment method",
			fields: fields{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindCreate),
					NewSubsIntent(SubsKindNew),
				},
				err: nil,
			},
			args: args{
				m: enum.PayMethodStripe,
			},
			want: CheckoutIntent{
				OneTimeKind: enum.OrderKindNull,
				SubsKind:    SubsKindNew,
				PayMethods: []enum.PayMethod{
					enum.PayMethodStripe,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coi := CheckoutIntents{
				intents: tt.fields.intents,
				err:     tt.fields.err,
			}
			got, err := coi.Get(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCheckoutIntents(t *testing.T) {
	type args struct {
		m reader.Membership
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
				m: reader.Membership{},
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindCreate),
					NewSubsIntent(SubsKindNew),
				},
				err: nil,
			},
		},
		{
			name: "Invalid stripe",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithSubsStatus(enum.SubsStatusUnpaid).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindCreate),
					NewSubsIntent(SubsKindNew),
				},
				err: nil,
			},
		},
		{
			name: "Onetime renewal",
			args: args{
				m: reader.NewMockMemberBuilder("").Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindRenew),
					NewSubsIntent(SubsKindOneTimeToStripe),
				},
				err: nil,
			},
		},
		{
			name: "Onetime upgrade",
			args: args{
				m: reader.NewMockMemberBuilder("").Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindUpgrade),
					NewSubsIntent(SubsKindOneTimeToStripe),
				},
				err: nil,
			},
		},
		{
			name: "Onetime premium buy addon",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe can purchase standard addon",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe can purchase premium addon",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can upgrade",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewSubsIntent(SubsKindUpgrade),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can purchase addon of same tier",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "Stripe standard can switch cycle of different cycle or purchase addon of same tier",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdMonthEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
					NewSubsIntent(SubsKindSwitchCycle),
				},
				err: nil,
			},
		},
		{
			name: "IAP cannot upgrade",
			args: args{
				m: reader.NewMockMemberBuilder("").
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
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PriceStdYear.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "IAP premium can purchase premium addon",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.PremiumEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
				},
				err: nil,
			},
		},
		{
			name: "IAP premium can purchase standard addon",
			args: args{
				m: reader.NewMockMemberBuilder("").
					WithPrice(faker.PricePrm.Price).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.StdYearEdition,
			},
			want: CheckoutIntents{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindAddOn),
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
