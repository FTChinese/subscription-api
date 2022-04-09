package reader

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"reflect"
	"testing"
)

func TestFtcPrice_ApplicableOffer(t *testing.T) {

	type args struct {
		enjoys []price.OfferKind
	}
	tests := []struct {
		name   string
		fields PaywallPrice
		args   args
		want   price.Discount
	}{
		{
			name:   "New member no offer",
			fields: MockPwPriceStdYear,
			args: args{
				enjoys: []price.OfferKind{
					price.OfferKindPromotion,
				},
			},
			want: price.Discount{},
		},
		{
			name:   "Existing member enjoys retention offer",
			fields: MockPwPriceStdYear,
			args: args{
				enjoys: []price.OfferKind{
					price.OfferKindPromotion,
					price.OfferKindRetention,
				},
			},
			want: MockPwPriceStdYear.Offers[1],
		},
		{
			name:   "Expired member enjoys win-back offer",
			fields: MockPwPriceStdYear,
			args: args{
				enjoys: []price.OfferKind{
					price.OfferKindPromotion,
					price.OfferKindWinBack,
				},
			},
			want: MockPwPriceStdYear.Offers[2],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := PaywallPrice{
				FtcPrice: tt.fields.FtcPrice,
				Offers:   tt.fields.Offers,
			}

			got := f.Offers.FindApplicable(tt.args.enjoys)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicableOffer() = \n%v, want \n%v", got, tt.want)
				return
			}

			t.Logf("Offer %s", faker.MustMarshalIndent(got))
		})
	}
}
