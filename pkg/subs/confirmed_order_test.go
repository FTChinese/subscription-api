package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewConfirmedOrder(t *testing.T) {
	orderCreate := NewMockOrderBuilder("").Build()
	payResultForCreate := MockNewPaymentResult(orderCreate)

	orderRenewal := NewMockOrderBuilder("").WithKind(enum.OrderKindRenew).Build()
	payResultForRenewal := MockNewPaymentResult(orderRenewal)

	orderUpgrade := NewMockOrderBuilder("").WithKind(enum.OrderKindUpgrade).Build()
	payResultForUpgrade := MockNewPaymentResult(orderUpgrade)

	orderAddOn := NewMockOrderBuilder("").WithKind(enum.OrderKindAddOn).Build()
	payResultForAddOn := MockNewPaymentResult(orderAddOn)

	member := reader.NewMockMemberBuilder(orderUpgrade.FtcID.String).
		WithExpiration(time.Now().AddDate(0, 3, 0)).
		Build()

	type args struct {
		params ConfirmationParams
	}
	tests := []struct {
		name    string
		args    args
		want    ConfirmedOrder
		wantErr bool
	}{
		{
			name: "Confirm order for new subscription",
			args: args{
				params: ConfirmationParams{
					Payment: payResultForCreate,
					Order:   orderCreate,
					Member:  reader.Membership{},
				},
			},
			want: ConfirmedOrder{
				Order: Order{
					ID:            orderCreate.ID,
					MemberID:      orderCreate.MemberID,
					PlanID:        orderCreate.PlanID,
					DiscountID:    orderCreate.DiscountID,
					Price:         orderCreate.Price,
					Edition:       orderCreate.Edition,
					Charge:        orderCreate.Charge,
					Kind:          orderCreate.Kind,
					PaymentMethod: orderCreate.PaymentMethod,
					WxAppID:       orderCreate.WxAppID,
					CreatedAt:     orderCreate.CreatedAt,
					ConfirmedAt:   payResultForCreate.ConfirmedUTC,
					DateRange: dt.DateRange{
						StartDate: chrono.DateFrom(payResultForCreate.ConfirmedUTC.Time),
						EndDate:   chrono.DateFrom(payResultForCreate.ConfirmedUTC.AddDate(1, 0, 1)),
					},
					LiveMode: orderCreate.LiveMode,
				},
				AddOn: addon.AddOn{},
			},
			wantErr: false,
		},
		{
			name: "Confirm order for renewal",
			args: args{
				params: ConfirmationParams{
					Payment: payResultForRenewal,
					Order:   orderRenewal,
					Member:  member,
				},
			},
			want: ConfirmedOrder{
				Order: Order{
					ID:            orderRenewal.ID,
					MemberID:      orderRenewal.MemberID,
					PlanID:        orderRenewal.PlanID,
					DiscountID:    orderRenewal.DiscountID,
					Price:         orderRenewal.Price,
					Edition:       orderRenewal.Edition,
					Charge:        orderRenewal.Charge,
					Kind:          orderRenewal.Kind,
					PaymentMethod: orderRenewal.PaymentMethod,
					WxAppID:       orderRenewal.WxAppID,
					CreatedAt:     orderRenewal.CreatedAt,
					ConfirmedAt:   payResultForRenewal.ConfirmedUTC,
					DateRange: dt.DateRange{
						StartDate: chrono.DateFrom(member.ExpireDate.Time),
						EndDate:   chrono.DateFrom(member.ExpireDate.AddDate(1, 0, 1)),
					},
					LiveMode: orderRenewal.LiveMode,
				},
				AddOn: addon.AddOn{},
			},
			wantErr: false,
		},
		{
			name: "Confirm order for upgrade",
			args: args{
				params: ConfirmationParams{
					Payment: payResultForUpgrade,
					Order:   orderUpgrade,
					Member:  member,
				},
			},
			want: ConfirmedOrder{
				Order: Order{
					ID:            orderUpgrade.ID,
					MemberID:      orderUpgrade.MemberID,
					PlanID:        orderUpgrade.PlanID,
					DiscountID:    orderUpgrade.DiscountID,
					Price:         orderUpgrade.Price,
					Edition:       orderUpgrade.Edition,
					Charge:        orderUpgrade.Charge,
					Kind:          orderUpgrade.Kind,
					PaymentMethod: orderUpgrade.PaymentMethod,
					WxAppID:       orderUpgrade.WxAppID,
					CreatedAt:     orderUpgrade.CreatedAt,
					ConfirmedAt:   payResultForUpgrade.ConfirmedUTC,
					DateRange: dt.DateRange{
						StartDate: chrono.DateFrom(payResultForUpgrade.ConfirmedUTC.Time),
						EndDate:   chrono.DateFrom(payResultForUpgrade.ConfirmedUTC.AddDate(1, 0, 1)),
					},
					LiveMode: orderUpgrade.LiveMode,
				},
				AddOn: addon.AddOn{
					ID:              "",
					Edition:         member.Edition,
					CycleCount:      0,
					DaysRemained:    member.RemainingDays(),
					CarryOverSource: addon.CarryOverFromUpgrade,
					PaymentMethod:   member.PaymentMethod,
					CompoundID:      member.CompoundID,
					OrderID:         null.StringFrom(orderUpgrade.ID),
					PlanID:          member.FtcPlanID,
					CreatedUTC:      chrono.TimeNow(),
					ConsumedUTC:     chrono.Time{},
				},
			},
			wantErr: false,
		},
		{
			name: "Confirm order for add-on",
			args: args{
				params: ConfirmationParams{
					Payment: payResultForAddOn,
					Order:   orderAddOn,
					Member:  member,
				},
			},
			want: ConfirmedOrder{
				Order: Order{
					ID:            orderAddOn.ID,
					MemberID:      orderAddOn.MemberID,
					PlanID:        orderAddOn.PlanID,
					DiscountID:    orderAddOn.DiscountID,
					Price:         orderAddOn.Price,
					Edition:       orderAddOn.Edition,
					Charge:        orderAddOn.Charge,
					Kind:          orderAddOn.Kind,
					PaymentMethod: orderAddOn.PaymentMethod,
					WxAppID:       orderAddOn.WxAppID,
					CreatedAt:     orderAddOn.CreatedAt,
					ConfirmedAt:   payResultForAddOn.ConfirmedUTC,
					DateRange:     dt.DateRange{},
					LiveMode:      orderAddOn.LiveMode,
				},
				AddOn: orderAddOn.ToAddOn(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfirmedOrder(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfirmedOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.AddOn.IsZero() {
				tt.want.AddOn.ID = got.AddOn.ID
				tt.want.AddOn.CreatedUTC = got.AddOn.CreatedUTC
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfirmedOrder() \ngot = %v, \nwant %v", got, tt.want)
			}
		})
	}
}
