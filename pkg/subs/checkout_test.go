package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var account = reader.FtcAccount{
	FtcID:    uuid.New().String(),
	UnionID:  null.String{},
	StripeID: null.String{},
	Email:    "any@example.org",
	UserName: null.StringFrom("World"),
	VIP:      false,
}

var planStdYear = product.ExpandedPlan{
	Plan: product.Plan{
		ID:        "plan_MynUQDQY1TSQ",
		ProductID: "prod_zjWdiTUpDN8l",
		Price:     258,
		Edition: product.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Description: "",
	},
	Discount: product.Discount{
		DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		PriceOff: null.FloatFrom(40),
		Percent:  null.Int{},
		DateTimeRange: dt.DateTimeRange{
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		},
		Description: null.String{},
	},
}

var planPrmYear = product.ExpandedPlan{
	Plan: product.Plan{
		ID:        "plan_vRUzRQ3aglea",
		ProductID: "prod_IaoK5SbK79g8",
		Price:     1998,
		Edition: product.Edition{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		Description: "",
	},
	Discount: product.Discount{
		DiscID:   null.StringFrom("dsc_a1Vp92cfFAih"),
		PriceOff: null.FloatFrom(300),
		Percent:  null.Int{},
		DateTimeRange: dt.DateTimeRange{
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		},
		Description: null.String{},
	},
}

func TestNewCheckedItem(t *testing.T) {
	type args struct {
		ep product.ExpandedPlan
	}
	tests := []struct {
		name string
		args args
		want CheckedItem
	}{
		{
			name: "Checkout Standard Edition",
			args: args{
				ep: planStdYear,
			},
			want: CheckedItem{
				Plan:     planStdYear.Plan,
				Discount: planStdYear.Discount,
			},
		},
		{
			name: "Checkout Premium Edition",
			args: args{
				ep: planPrmYear,
			},
			want: CheckedItem{
				Plan:     planPrmYear.Plan,
				Discount: planPrmYear.Discount,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCheckedItem(tt.args.ep)
			assert.Equal(t, got, tt.want)
			t.Logf("Checkout item %+v", got)
		})
	}
}

func TestPaymentConfig_Checkout(t *testing.T) {
	type fields struct {
		dryRun  bool
		Account reader.FtcAccount
		Plan    product.ExpandedPlan
		Method  enum.PayMethod
		WxAppID null.String
	}
	type args struct {
		bs   []BalanceSource
		kind enum.OrderKind
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Checkout
	}{
		{
			name: "New standard order",
			fields: fields{
				dryRun:  false,
				Account: account,
				Plan:    planStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				bs:   nil,
				kind: enum.OrderKindCreate,
			},
			want: Checkout{
				Kind: enum.OrderKindCreate,
				Item: CheckedItem{
					Plan:     planStdYear.Plan,
					Discount: planStdYear.Discount,
				},
				Wallet: Wallet{
					Balance:   0,
					CreatedAt: chrono.TimeNow(),
					Sources:   nil,
				},
				Duration: product.Duration{
					CycleCount: 1,
					ExtraDays:  1,
				},
				Payable: product.Charge{
					Amount:   planStdYear.Price - planStdYear.Discount.PriceOff.Float64,
					Currency: "cny",
				},
				IsFree:   false,
				LiveMode: true,
			},
		},
		{
			name: "New premium order",
			fields: fields{
				dryRun:  false,
				Account: account,
				Plan:    planPrmYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				bs:   nil,
				kind: enum.OrderKindCreate,
			},
			want: Checkout{
				Kind: enum.OrderKindCreate,
				Item: CheckedItem{
					Plan:     planPrmYear.Plan,
					Discount: planPrmYear.Discount,
				},
				Duration: product.Duration{
					CycleCount: 1,
					ExtraDays:  1,
				},
				Payable: product.Charge{
					Amount:   planPrmYear.Price - planPrmYear.Discount.PriceOff.Float64,
					Currency: "cny",
				},
				IsFree:   false,
				LiveMode: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PaymentConfig{
				DryRun:  tt.fields.dryRun,
				Account: tt.fields.Account,
				Plan:    tt.fields.Plan,
				Method:  tt.fields.Method,
				WxAppID: tt.fields.WxAppID,
			}
			got := c.Checkout(tt.args.bs, tt.args.kind)
			assert.Equal(t, got.Kind, tt.want.Kind)
			assert.Equal(t, got.Item, tt.want.Item)
			assert.Equal(t, got.Wallet.Balance, tt.want.Wallet.Balance)
			assert.Equal(t, got.Duration, tt.want.Duration)
			assert.Equal(t, got.Payable, tt.want.Payable)
			assert.Equal(t, got.IsFree, tt.want.IsFree)
			assert.Equal(t, got.LiveMode, tt.want.LiveMode)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestPaymentConfig_BuildOrder(t *testing.T) {
	type fields struct {
		dryRun  bool
		Account reader.FtcAccount
		Plan    product.ExpandedPlan
		Method  enum.PayMethod
		WxAppID null.String
	}
	type args struct {
		checkout Checkout
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Order
		wantErr bool
	}{
		{
			name: "New order",
			fields: fields{
				dryRun:  false,
				Account: account,
				Plan:    planStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				checkout: Checkout{
					Kind: enum.OrderKindCreate,
					Item: CheckedItem{
						Plan:     planStdYear.Plan,
						Discount: planStdYear.Discount,
					},
					Wallet: Wallet{
						Balance:   0,
						CreatedAt: chrono.TimeNow(),
						Sources:   nil,
					},
					Duration: product.Duration{
						CycleCount: 1,
						ExtraDays:  1,
					},
					Payable: product.Charge{
						Amount:   128,
						Currency: "cny",
					},
					IsFree:   false,
					LiveMode: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PaymentConfig{
				DryRun:  tt.fields.dryRun,
				Account: tt.fields.Account,
				Plan:    tt.fields.Plan,
				Method:  tt.fields.Method,
				WxAppID: tt.fields.WxAppID,
			}
			got, err := c.BuildOrder(tt.args.checkout)
			assert.NoError(t, err)

			assert.NotEmpty(t, got.ID)
			assert.NotZero(t, got.Price)
			assert.NotZero(t, got.Amount)
			assert.Zero(t, got.ConfirmedAt)
		})
	}
}

func TestPaymentConfig_BuildIntent(t *testing.T) {
	type fields struct {
		dryRun  bool
		Account reader.FtcAccount
		Plan    product.ExpandedPlan
		Method  enum.PayMethod
		WxAppID null.String
	}
	type args struct {
		bs   []BalanceSource
		kind enum.OrderKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PaymentIntent
		wantErr bool
	}{
		{
			name: "New order",
			fields: fields{
				dryRun:  false,
				Account: account,
				Plan:    planStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				bs:   nil,
				kind: enum.OrderKindCreate,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PaymentConfig{
				DryRun:  tt.fields.dryRun,
				Account: tt.fields.Account,
				Plan:    tt.fields.Plan,
				Method:  tt.fields.Method,
				WxAppID: tt.fields.WxAppID,
			}
			got, err := c.BuildIntent(tt.args.bs, tt.args.kind)
			assert.NoError(t, err)
			assert.NotZero(t, got.Checkout)
		})
	}
}

func TestPaymentConfig_UpgradeIntent(t *testing.T) {

	config := NewPayment(account, planPrmYear)
	mmb := reader.Membership{
		MemberID: reader.MemberID{
			FtcID: null.StringFrom(uuid.New().String()),
		}.MustNormalize(),
		Edition:       product.NewStdYearEdition(),
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
		PaymentMethod: enum.PayMethodWx,
		FtcPlanID:     null.StringFrom(planStdYear.Plan.ID),
	}

	type args struct {
		bs []BalanceSource
		m  reader.Membership
	}
	tests := []struct {
		name     string
		args     args
		wantFree bool
		wantErr  bool
	}{
		{
			name: "Free upgrade",
			args: args{
				bs: MockBalanceSourceN(10),
				m:  mmb,
			},
			wantFree: true,
			wantErr:  false,
		},
		{
			name: "Not free upgrade",
			args: args{
				bs: MockBalanceSourceN(3),
				m:  mmb,
			},
			wantFree: false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := config.UpgradeIntent(tt.args.bs, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeIntent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.IsFree != tt.wantFree {
				t.Errorf("UpradeIntent() IsFree = %t, wantFree %t", got.IsFree, tt.wantFree)
			}

			t.Logf("Wallet: %+v", got.Wallet)
			t.Logf("Payable: %+v", got.Payable)
			t.Logf("Duration: %+v", got.Duration)
			t.Logf("Is free: %t", got.IsFree)
			t.Logf("Order: %v", got.Result.Order)
			t.Logf("New member: %v", got.Result.Membership)
		})
	}
}
