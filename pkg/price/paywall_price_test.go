package price

import (
	"github.com/FTChinese/subscription-api/faker"
	"reflect"
	"testing"
)

func TestFtcPrice_ApplicableOffer(t *testing.T) {

	type args struct {
		enjoys []OfferKind
	}
	tests := []struct {
		name   string
		fields PaywallPrice
		args   args
		want   Discount
	}{
		{
			name:   "New member no offer",
			fields: MockPriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
				},
			},
			want: Discount{},
		},
		{
			name:   "Existing member enjoys retention offer",
			fields: MockPriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
					OfferKindRetention,
				},
			},
			want: MockPriceStdYear.Offers[1],
		},
		{
			name:   "Expired member enjoys win-back offer",
			fields: MockPriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
					OfferKindWinBack,
				},
			},
			want: MockPriceStdYear.Offers[2],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := PaywallPrice{
				Price:  tt.fields.Price,
				Offers: tt.fields.Offers,
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
