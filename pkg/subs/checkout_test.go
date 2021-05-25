package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"reflect"
	"testing"
	"time"
)

func TestNewCheckout(t *testing.T) {
	type args struct {
		ftcPrice price.FtcPrice
		m        reader.Membership
	}
	tests := []struct {
		name    string
		args    args
		want    Checkout
		wantErr bool
	}{
		{
			name: "Create",
			args: args{
				ftcPrice: price.MockPriceStdYear,
				m:        reader.Membership{},
			},
			want: Checkout{
				Kind:     enum.OrderKindCreate,
				Price:    price.MockPriceStdYear.Price,
				Offer:    price.Discount{},
				LiveMode: true,
			},
		},
		{
			name: "Renew",
			args: args{
				ftcPrice: price.MockPriceStdYear,
				m:        reader.NewMockMemberBuilder("").Build(),
			},
			want: Checkout{
				Kind:     enum.OrderKindRenew,
				Price:    price.MockPriceStdYear.Price,
				Offer:    price.FtcOffers[price.StdYearEdition][0],
				LiveMode: true,
			},
		},
		{
			name: "Win-back",
			args: args{
				ftcPrice: price.MockPriceStdYear,
				m: reader.NewMockMemberBuilder("").
					WithExpiration(time.Now().AddDate(0, -1, 0)).
					Build(),
			},
			want: Checkout{
				Kind:     enum.OrderKindCreate,
				Price:    price.MockPriceStdYear.Price,
				Offer:    price.FtcOffers[price.StdYearEdition][1],
				LiveMode: true,
			},
		},
		{
			name: "Upgrade using retention offer",
			args: args{
				ftcPrice: price.MockPricePrm,
				m:        reader.NewMockMemberBuilder("").Build(),
			},
			want: Checkout{
				Kind:     enum.OrderKindUpgrade,
				Price:    price.MockPricePrm.Price,
				Offer:    price.FtcOffers[price.PremiumEdition][0],
				LiveMode: true,
			},
		},
		{
			name: "Upgrade using win-back offer",
			args: args{
				ftcPrice: price.MockPricePrm,
				m: reader.NewMockMemberBuilder("").
					WithExpiration(time.Now().AddDate(0, -1, 0)).
					Build(),
			},
			want: Checkout{
				Kind:     enum.OrderKindCreate,
				Price:    price.MockPricePrm.Price,
				Offer:    price.FtcOffers[price.PremiumEdition][1],
				LiveMode: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCheckout(tt.args.ftcPrice, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCheckout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCheckout() got = %v, want %v", got, tt.want)
			}
		})
	}
}
