package price

import (
	"github.com/FTChinese/subscription-api/faker"
	"reflect"
	"testing"
)

func TestDiscountListJSON_FindApplicable(t *testing.T) {
	type args struct {
		filters []OfferKind
	}
	tests := []struct {
		name string
		l    DiscountListJSON
		args args
		want Discount
	}{
		{
			name: "Find applicable offer",
			l:    MockPriceStdYear.Offers,
			args: args{
				filters: []OfferKind{
					OfferKindRetention,
				},
			},
			want: MockPriceStdYear.Offers[1],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.l.FindApplicable(tt.args.filters)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindApplicable() = %v, want %v", got, tt.want)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
