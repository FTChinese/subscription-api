package cart

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewFtcCart(t *testing.T) {
	now := time.Now()
	type args struct {
		ftcPrice price.FtcPrice
	}
	tests := []struct {
		name string
		args args
		want Cart
	}{
		{
			name: "Price with discount",
			args: args{
				ftcPrice: price.FtcPrice{
					Original: price.Price{
						ID: "plan_MynUQDQY1TSQ",
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						Active:     true,
						Currency:   price.CurrencyCNY,
						LiveMode:   true,
						Nickname:   null.String{},
						ProductID:  "prod_zjWdiTUpDN8l",
						Source:     price.SourceFTC,
						UnitAmount: 298,
					},
					PromotionOffer: price.Discount{
						DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
						PriceOff: null.FloatFrom(130),
						Percent:  null.Int{},
						DateTimePeriod: dt.DateTimePeriod{
							StartUTC: chrono.TimeFrom(now),
							EndUTC:   chrono.TimeFrom(now.AddDate(0, 0, 2)),
						},
						Description: null.String{},
					},
				},
			},
			want: Cart{
				Price: price.Price{
					ID: "plan_MynUQDQY1TSQ",
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Active:     true,
					Currency:   price.CurrencyCNY,
					LiveMode:   true,
					Nickname:   null.String{},
					ProductID:  "prod_zjWdiTUpDN8l",
					Source:     price.SourceFTC,
					UnitAmount: 298,
				},
				Discount: price.Discount{
					DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
					PriceOff: null.FloatFrom(130),
					Percent:  null.Int{},
					DateTimePeriod: dt.DateTimePeriod{
						StartUTC: chrono.TimeFrom(now),
						EndUTC:   chrono.TimeFrom(now.AddDate(0, 0, 2)),
					},
					Description: null.String{},
				},
			},
		},
		{
			name: "Price without discount",
			args: args{
				ftcPrice: price.FtcPrice{
					Original: price.Price{
						ID: "plan_MynUQDQY1TSQ",
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						Active:     true,
						Currency:   price.CurrencyCNY,
						LiveMode:   true,
						Nickname:   null.String{},
						ProductID:  "prod_zjWdiTUpDN8l",
						Source:     price.SourceFTC,
						UnitAmount: 298,
					},
					PromotionOffer: price.Discount{},
				},
			},
			want: Cart{
				Price: price.Price{
					ID: "plan_MynUQDQY1TSQ",
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Active:     true,
					Currency:   price.CurrencyCNY,
					LiveMode:   true,
					Nickname:   null.String{},
					ProductID:  "prod_zjWdiTUpDN8l",
					Source:     price.SourceFTC,
					UnitAmount: 298,
				},
				Discount: price.Discount{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFtcCart(tt.args.ftcPrice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFtcCart() = %v, want %v", got, tt.want)
			}
		})
	}
}
