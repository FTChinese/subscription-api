package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewFtcCart(t *testing.T) {
	now := time.Now()
	type args struct {
		ftcPrice FtcPrice
	}
	tests := []struct {
		name string
		args args
		want Cart
	}{
		{
			name: "Price with discount",
			args: args{
				ftcPrice: FtcPrice{
					Price: Price{
						ID: "plan_MynUQDQY1TSQ",
						Edition: Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						Active:     true,
						Currency:   CurrencyCNY,
						LiveMode:   true,
						Nickname:   null.String{},
						ProductID:  "prod_zjWdiTUpDN8l",
						Source:     SourceFTC,
						UnitAmount: 298,
					},
					PromotionOffer: Discount{
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
				Price: Price{
					ID: "plan_MynUQDQY1TSQ",
					Edition: Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Active:     true,
					Currency:   CurrencyCNY,
					LiveMode:   true,
					Nickname:   null.String{},
					ProductID:  "prod_zjWdiTUpDN8l",
					Source:     SourceFTC,
					UnitAmount: 298,
				},
				Discount: Discount{
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
				ftcPrice: FtcPrice{
					Price: Price{
						ID: "plan_MynUQDQY1TSQ",
						Edition: Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						Active:     true,
						Currency:   CurrencyCNY,
						LiveMode:   true,
						Nickname:   null.String{},
						ProductID:  "prod_zjWdiTUpDN8l",
						Source:     SourceFTC,
						UnitAmount: 298,
					},
					PromotionOffer: Discount{},
				},
			},
			want: Cart{
				Price: Price{
					ID: "plan_MynUQDQY1TSQ",
					Edition: Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Active:     true,
					Currency:   CurrencyCNY,
					LiveMode:   true,
					Nickname:   null.String{},
					ProductID:  "prod_zjWdiTUpDN8l",
					Source:     SourceFTC,
					UnitAmount: 298,
				},
				Discount: Discount{},
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
