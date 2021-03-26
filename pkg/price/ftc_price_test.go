package price

import (
	"reflect"
	"testing"
)

func TestFtcPrice_ApplicableOffer(t *testing.T) {

	type args struct {
		enjoys []OfferKind
	}
	tests := []struct {
		name   string
		fields FtcPrice
		args   args
		want   Discount
	}{
		{
			name:   "New member no offer",
			fields: PriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
				},
			},
			want: Discount{},
		},
		{
			name:   "Existing member enjoys retention offer",
			fields: PriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
					OfferKindRetention,
				},
			},
			want: FtcOffers[StdYearEdition][0],
		},
		{
			name:   "Expired member enjoys win-back offer",
			fields: PriceStdYear,
			args: args{
				enjoys: []OfferKind{
					OfferKindPromotion,
					OfferKindWinBack,
				},
			},
			want: FtcOffers[StdYearEdition][1],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FtcPrice{
				Price:  tt.fields.Price,
				Offers: tt.fields.Offers,
			}
			if got := f.ApplicableOffer(tt.args.enjoys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicableOffer() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}