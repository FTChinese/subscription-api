package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestEnv_CreateOrder(t *testing.T) {
	wxID := faker.GenWxID()
	repo := test.NewRepo()

	p := test.NewPersona()
	repo.MustSaveMembership(p.Membership())

	p2 := test.NewPersona().SetPayMethod(enum.PayMethodApple)
	repo.MustSaveMembership(p2.Membership())

	type fields struct {
		rwdDB  *sqlx.DB
		logger *zap.Logger
	}
	type args struct {
		config subs.PaymentConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    subs.PaymentIntent
		wantErr bool
	}{
		{
			name: "New order",
			fields: fields{
				rwdDB:  test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.PaymentConfig{
					Account: p.FtcAccount(),
					Price:   faker.PriceStdYear,
					Method:  enum.PayMethodAli,
					WxAppID: null.String{},
				},
			},
			want: subs.PaymentIntent{
				Checkout: subs.Checkout{
					Kind:     enum.OrderKindCreate,
					Item:     cart.NewCheckoutItem(faker.PriceStdYear),
					Payable:  cart.NewCheckoutItem(faker.PriceStdYear).Payable(),
					LiveMode: true,
				},
				Order: subs.Order{
					ID:         "",
					MemberID:   p.AccountID(),
					PlanID:     faker.PriceStdYear.Original.ID,
					DiscountID: null.String{},
					Price:      faker.PriceStdYear.Original.UnitAmount,
					Edition:    faker.PriceStdYear.Original.Edition,
					Charge: price.Charge{
						Amount:   faker.PriceStdYear.Original.UnitAmount,
						Currency: "cny",
					},
					Kind:          enum.OrderKindCreate,
					PaymentMethod: enum.PayMethodAli,
					WxAppID:       null.String{},
					CreatedAt:     chrono.Time{},
					ConfirmedAt:   chrono.Time{},
					DateRange:     dt.DateRange{},
					LiveMode:      true,
				},
			},
			wantErr: false,
		},
		{
			name: "Renewal order",
			fields: fields{
				rwdDB:  test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.PaymentConfig{
					Account: p.FtcAccount(),
					Price:   faker.PriceStdYear,
					Method:  enum.PayMethodWx,
					WxAppID: null.StringFrom(wxID),
				},
			},
			want: subs.PaymentIntent{
				Checkout: subs.Checkout{
					Kind:     enum.OrderKindRenew,
					Item:     cart.NewCheckoutItem(faker.PriceStdYear),
					Payable:  cart.NewCheckoutItem(faker.PriceStdYear).Payable(),
					LiveMode: true,
				},
				Order: subs.Order{
					ID:         "",
					MemberID:   p.AccountID(),
					PlanID:     faker.PriceStdYear.Original.ID,
					DiscountID: null.String{},
					Price:      faker.PriceStdYear.Original.UnitAmount,
					Edition:    faker.PriceStdYear.Original.Edition,
					Charge: price.Charge{
						Amount:   faker.PriceStdYear.Original.UnitAmount,
						Currency: "cny",
					},
					Kind:          enum.OrderKindRenew,
					PaymentMethod: enum.PayMethodWx,
					WxAppID:       null.StringFrom(wxID),
					CreatedAt:     chrono.Time{},
					ConfirmedAt:   chrono.Time{},
					DateRange:     dt.DateRange{},
					LiveMode:      true,
				},
			},
			wantErr: false,
		},
		{
			name: "Upgrade order",
			fields: fields{
				rwdDB:  test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.PaymentConfig{
					Account: p.FtcAccount(),
					Price:   faker.PricePrm,
					Method:  enum.PayMethodWx,
					WxAppID: null.StringFrom(wxID),
				},
			},
			want: subs.PaymentIntent{
				Checkout: subs.Checkout{
					Kind:     enum.OrderKindUpgrade,
					Item:     cart.NewCheckoutItem(faker.PricePrm),
					Payable:  cart.NewCheckoutItem(faker.PricePrm).Payable(),
					LiveMode: true,
				},
				Order: subs.Order{
					ID:         "",
					MemberID:   p.AccountID(),
					PlanID:     faker.PricePrm.Original.ID,
					DiscountID: null.String{},
					Price:      faker.PricePrm.Original.UnitAmount,
					Edition:    faker.PricePrm.Original.Edition,
					Charge: price.Charge{
						Amount:   faker.PricePrm.Original.UnitAmount,
						Currency: "cny",
					},
					Kind:          enum.OrderKindUpgrade,
					PaymentMethod: enum.PayMethodWx,
					WxAppID:       null.StringFrom(wxID),
					CreatedAt:     chrono.Time{},
					ConfirmedAt:   chrono.Time{},
					DateRange:     dt.DateRange{},
					LiveMode:      true,
				},
			},
			wantErr: false,
		},
		{
			name: "Add-on order",
			fields: fields{
				rwdDB:  test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.PaymentConfig{
					Account: p2.FtcAccount(),
					Price:   faker.PriceStdYear,
					Method:  enum.PayMethodWx,
					WxAppID: null.StringFrom(wxID),
				},
			},
			want: subs.PaymentIntent{
				Checkout: subs.Checkout{
					Kind:     enum.OrderKindAddOn,
					Item:     cart.NewCheckoutItem(faker.PriceStdYear),
					Payable:  cart.NewCheckoutItem(faker.PriceStdYear).Payable(),
					LiveMode: true,
				},
				Order: subs.Order{
					ID:         "",
					MemberID:   p2.AccountID(),
					PlanID:     faker.PriceStdYear.Original.ID,
					DiscountID: null.String{},
					Price:      faker.PriceStdYear.Original.UnitAmount,
					Edition:    faker.PriceStdYear.Original.Edition,
					Charge: price.Charge{
						Amount:   faker.PriceStdYear.Original.UnitAmount,
						Currency: "cny",
					},
					Kind:          enum.OrderKindAddOn,
					PaymentMethod: enum.PayMethodWx,
					WxAppID:       null.StringFrom(wxID),
					CreatedAt:     chrono.Time{},
					ConfirmedAt:   chrono.Time{},
					DateRange:     dt.DateRange{},
					LiveMode:      true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				rwdDB:  tt.fields.rwdDB,
				logger: tt.fields.logger,
			}
			got, err := env.CreateOrder(tt.args.config)

			tt.want.Order.ID = got.Order.ID
			tt.want.Order.CreatedAt = got.Order.CreatedAt

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateOrder() got = %v\n, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_LogOrderMeta(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		cache  *cache.Cache
		logger *zap.Logger
	}
	type args struct {
		m subs.OrderMeta
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Log order meta data",
			fields: fields{
				db: test.DB,
			},
			args: args{
				m: subs.OrderMeta{
					OrderID: db.MustOrderID(),
					Client:  faker.RandomClientApp(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				rwdDB:  tt.fields.db,
				logger: tt.fields.logger,
			}
			if err := env.LogOrderMeta(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("LogOrderMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveOrder(t *testing.T) {
	p := test.NewPersona()
	order := p.NewOrder(enum.OrderKindCreate)

	repo := test.NewRepo()
	repo.MustSaveOrder(order)

	type fields struct {
		db     *sqlx.DB
		cache  *cache.Cache
		logger *zap.Logger
	}
	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				orderID: order.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				rwdDB:  tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := env.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Order %+v", got)
		})
	}
}

func TestEnv_LoadFullOrder(t *testing.T) {

	p := test.NewPersona()
	order := p.NewOrder(enum.OrderKindCreate)

	t.Logf("Order id: %s", order.ID)

	test.NewRepo().MustSaveOrder(order)

	env := NewEnv(test.DB, zaptest.NewLogger(t))

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load full order",
			args: args{
				orderID: order.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.LoadFullOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFullOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)

			assert.NotZero(t, got.ID, order.ID)
		})
	}
}
