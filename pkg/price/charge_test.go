package price

import (
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestNewCharge(t *testing.T) {
	type args struct {
		price Price
		offer Discount
	}
	tests := []struct {
		name string
		args args
		want Charge
	}{
		{
			name: "Price with discount",
			args: args{
				price: MockPriceStdYear.Price,
				offer: Discount{
					DiscID:         null.StringFrom("id"),
					PriceOff:       null.FloatFrom(40),
					Percent:        null.IntFrom(85),
					DateTimePeriod: dt.DateTimePeriod{},
					Description:    null.String{},
					Kind:           OfferKindWinBack,
				},
			},
			want: Charge{
				Amount:   258,
				Currency: "cny",
			},
		},
		{
			name: "Price without discount",
			args: args{
				price: MockPriceStdYear.Price,
				offer: Discount{},
			},
			want: Charge{
				Amount:   298,
				Currency: "cny",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCharge(tt.args.price, tt.args.offer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCharge() = %v, want %v", got, tt.want)
			}
		})
	}
}
