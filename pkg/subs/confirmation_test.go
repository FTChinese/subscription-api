package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewPaymentConfirmed(t *testing.T) {
	now := time.Now()

	pc := NewPayment(reader.MockNewFtcAccount(enum.AccountKindFtc), faker.PriceStdYear).WithAlipay()
	co, _ := pc.checkout(reader.Membership{})

	order, err := pc.order(co)
	assert.Nil(t, err)

	order2, err := pc.order(co)
	assert.Nil(t, err)
	order2.ConfirmedAt = chrono.TimeFrom(now)
	order2.StartDate = chrono.DateFrom(now)
	order2.EndDate = chrono.DateFrom(now.AddDate(1, 0, 0))

	if err != nil {
		t.Error(err)
	}

	type args struct {
		p ConfirmationParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirm a new subscription",
			args: args{
				p: ConfirmationParams{
					Payment: PaymentResult{
						PaymentState:  ali.TradeStatusSuccess,
						Amount:        null.IntFrom(12800),
						TransactionID: "1234",
						OrderID:       order.ID,
						PaidAt:        chrono.TimeFrom(now),
						ConfirmedUTC:  chrono.TimeFrom(now),
						PayMethod:     enum.PayMethodAli,
					},
					Order:  order,
					Member: reader.Membership{},
				},
			},
			wantErr: false,
		},
		{
			name: "Already confirmed but out of sync",
			args: args{
				p: ConfirmationParams{
					Payment: PaymentResult{
						PaymentState:  ali.TradeStatusSuccess,
						Amount:        null.IntFrom(12800),
						TransactionID: "1234",
						OrderID:       order2.ID,
						PaidAt:        chrono.TimeFrom(now),
						ConfirmedUTC:  chrono.TimeFrom(now),
						PayMethod:     enum.PayMethodAli,
					},
					Order:  order2,
					Member: reader.Membership{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPaymentConfirmed(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaymentConfirmed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, now, got.Order.ConfirmedAt.Time)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestNewMembership(t *testing.T) {
	now := time.Now()

	orderCreate := MockOrder(
		faker.PriceStdYear, enum.OrderKindCreate).
		newOrRenewalConfirm(chrono.TimeFrom(now), chrono.Date{})

	orderRenew := MockOrder(faker.PriceStdYear, enum.OrderKindRenew).
		newOrRenewalConfirm(chrono.TimeFrom(now), chrono.DateFrom(now.AddDate(0, 1, 0)))

	orderUpgrade := MockOrder(faker.PricePrm, enum.OrderKindUpgrade).
		upgradeConfirm(chrono.TimeFrom(now))

	orderAddOn := MockOrder(faker.PriceStdYear, enum.OrderKindAddOn)
	orderAddOn.ConfirmedAt = chrono.TimeFrom(now)

	iapMember := reader.Membership{
		MemberID:      orderAddOn.MemberID,
		Edition:       orderAddOn.Edition,
		ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 1)),
		PaymentMethod: enum.PayMethodApple,
		AutoRenewal:   true,
		AppleSubsID:   null.StringFrom(faker.GenAppleSubID()),
	}.Sync()

	type args struct {
		p PaymentConfirmed
	}
	tests := []struct {
		name string
		args args
		want reader.Membership
	}{
		{
			name: "Build new membership",
			args: args{
				p: PaymentConfirmed{
					Order:    orderCreate,
					AddOn:    addon.AddOn{},
					Snapshot: reader.MemberSnapshot{},
				},
			},
			want: reader.Membership{
				MemberID:      orderCreate.MemberID,
				Edition:       orderCreate.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 1)),
				PaymentMethod: orderCreate.PaymentMethod,
				FtcPlanID:     null.StringFrom(orderCreate.PlanID),
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				ReservedDays:  addon.ReservedDays{},
			}.Sync(),
		},
		{
			name: "Build renewal membership",
			args: args{
				p: PaymentConfirmed{
					Order: orderRenew,
					AddOn: addon.AddOn{},
					Snapshot: reader.MemberSnapshot{
						SnapshotID: db.SnapshotID(),
						CreatedBy:  null.String{},
						CreatedUTC: chrono.TimeNow(),
						OrderID:    null.StringFrom(orderRenew.ID),
						Membership: reader.Membership{},
					},
				},
			},
			want: reader.Membership{
				MemberID:      orderRenew.MemberID,
				Edition:       orderRenew.Edition,
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 1, 1)),
				PaymentMethod: orderRenew.PaymentMethod,
				FtcPlanID:     null.StringFrom(orderRenew.PlanID),
			}.Sync(),
		},
		{
			name: "Build upgrade membership",
			args: args{
				p: PaymentConfirmed{
					Order: orderUpgrade,
					AddOn: addon.AddOn{
						ID:            db.AddOnID(),
						Edition:       price.StdYearEdition,
						CycleCount:    0,
						DaysRemained:  10,
						PaymentMethod: enum.PayMethodApple,
						OrderID:       null.StringFrom(orderUpgrade.ID),
						CompoundID:    orderUpgrade.CompoundID,
						CreatedUTC:    chrono.TimeNow(),
						ConsumedUTC:   chrono.Time{},
					},
					Snapshot: reader.MemberSnapshot{},
				},
			},
			want: reader.Membership{
				MemberID:      orderUpgrade.MemberID,
				Edition:       orderUpgrade.Edition,
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 1)),
				PaymentMethod: orderUpgrade.PaymentMethod,
				FtcPlanID:     null.StringFrom(orderUpgrade.PlanID),
				ReservedDays: addon.ReservedDays{
					Standard: 10,
				},
			}.Sync(),
		},
		{
			name: "Build addon membership",
			args: args{
				p: PaymentConfirmed{
					Order: orderAddOn,
					AddOn: orderAddOn.ToAddOn(),
					Snapshot: reader.MemberSnapshot{
						SnapshotID: db.SnapshotID(),
						CreatedBy:  null.String{},
						CreatedUTC: chrono.TimeNow(),
						OrderID:    null.StringFrom(orderAddOn.ID),
						Membership: iapMember,
					},
				},
			},
			want: iapMember.WithReservedDays(addon.ReservedDays{
				Standard: 367,
			}).Sync(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMembership(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMembership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfirmationResult(t *testing.T) {
	type args struct {
		p ConfirmationParams
	}
	tests := []struct {
		name    string
		args    args
		want    ConfirmationResult
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfirmationResult(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfirmationResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfirmationResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}
