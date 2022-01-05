package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewOrderInvoice(t *testing.T) {
	now := time.Now()
	userID := uuid.New().String()

	order := NewMockOrderBuilder(userID).
		Build()

	type args struct {
		timeParams PurchasedTimeParams
		o          Order
		p          price.Price
	}
	tests := []struct {
		name    string
		args    args
		want    invoice.Invoice
		wantErr bool
	}{
		{
			name: "Invoice from order",
			args: args{
				timeParams: PurchasedTimeParams{
					ConfirmedAt:    chrono.TimeFrom(now),
					ExpirationDate: chrono.Date{},
					PeriodCount: dt.YearMonthDay{
						Years:  1,
						Months: 0,
						Days:   1,
					},
					OrderKind: enum.OrderKindCreate,
				},
				o: order,
			},
			want: invoice.Invoice{
				ID:         "",
				CompoundID: userID,
				Edition:    price.StdYearEdition,
				YearMonthDay: dt.YearMonthDay{
					Years:  1,
					Months: 0,
					Days:   1,
				},
				AddOnSource:   "",
				OrderID:       null.StringFrom(order.ID),
				OrderKind:     enum.OrderKindCreate,
				PaidAmount:    order.PayableAmount,
				PaymentMethod: order.PaymentMethod,
				CreatedUTC:    chrono.TimeNow(),
				ConsumedUTC:   chrono.TimeNow(),
				ChronoPeriod: dt.ChronoPeriod{
					StartUTC: chrono.TimeFrom(now),
					EndUTC:   chrono.TimeFrom(now.AddDate(1, 0, 1)),
				},
				CarriedOverUtc: chrono.Time{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newOrderInvoice(tt.args.timeParams, tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("newOrderInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want.CreatedUTC = got.CreatedUTC
			tt.want.ConsumedUTC = got.ConsumedUTC
			tt.want.ID = got.ID

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newOrderInvoice() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestInvoices_membership(t *testing.T) {
	userID := uuid.New().String()
	current := reader.NewMockMemberBuilder(userID).Build()
	now := time.Now()

	type fields struct {
		Purchased   invoice.Invoice
		CarriedOver invoice.Invoice
	}
	type args struct {
		userID  ids.UserIDs
		current reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Create new membership",
			fields: fields{
				Purchased: invoice.NewMockInvoiceBuilder().
					WithFtcID(userID).
					SetPeriodStart(now).
					Build(),
				CarriedOver: invoice.Invoice{},
			},
			args: args{
				userID:  ids.NewFtcUserID(userID),
				current: reader.Membership{},
			},
			want: reader.Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       price.StdYearEdition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(price.MockPriceStdYear.ID),
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				AddOn:         addon.AddOn{},
			}.Sync(),
			wantErr: false,
		},
		{
			name: "Renew membership",
			fields: fields{
				Purchased: invoice.NewMockInvoiceBuilder().
					WithFtcID(userID).
					SetPeriodStart(current.ExpireDate.Time).
					WithOrderKind(enum.OrderKindRenew).
					Build(),
				CarriedOver: invoice.Invoice{},
			},
			args: args{
				userID:  ids.NewFtcUserID(userID),
				current: current,
			},
			want: reader.Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       price.StdYearEdition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(current.ExpireDate.AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(price.MockPriceStdYear.ID),
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				AddOn:         addon.AddOn{},
			}.Sync(),
			wantErr: false,
		},
		{
			name: "Upgrade membership",
			fields: fields{
				Purchased: invoice.NewMockInvoiceBuilder().
					WithFtcID(userID).
					SetPeriodStart(now).
					WithOrderKind(enum.OrderKindUpgrade).
					Build(),
				CarriedOver: current.CarryOverInvoice(),
			},
			args: args{
				userID:  ids.NewFtcUserID(userID),
				current: current,
			},
			want: reader.Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       price.StdYearEdition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(price.MockPriceStdYear.ID),
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				AddOn:         current.CarriedOverAddOn(),
			}.Sync(),
			wantErr: false,
		},
		{
			name: "Add on",
			fields: fields{
				Purchased: invoice.NewMockInvoiceBuilder().
					WithFtcID(userID).
					WithOrderKind(enum.OrderKindAddOn).
					Build(),
				CarriedOver: invoice.Invoice{},
			},
			args: args{
				userID:  ids.NewFtcUserID(userID),
				current: current,
			},
			want:    current.PlusAddOn(addon.New(enum.TierStandard, 367)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := Invoices{
				Purchased:   tt.fields.Purchased,
				CarriedOver: tt.fields.CarriedOver,
			}
			got, err := i.membership(tt.args.userID, tt.args.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("membership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("membership() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
