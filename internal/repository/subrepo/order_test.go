package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestEnv_CreateOrder(t *testing.T) {
	wxID := faker.GenWxID()
	repo := test.NewRepo()

	newPersona := test.NewPersona()
	renewalPerson := test.NewPersona()
	upgradePerson := test.NewPersona()
	addOnPerson := test.NewPersona()

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		counter subs.Counter
		p       *test.Persona
	}
	tests := []struct {
		name    string
		args    args
		want    subs.Order
		wantErr bool
	}{
		{
			name: "New order",
			args: args{
				counter: subs.Counter{
					BaseAccount: newPersona.BaseAccount(),
					FtcPrice:    price.PriceStdYear,
					Method:      enum.PayMethodAli,
					WxAppID:     null.String{},
				},
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    newPersona.AccountID(),
				PlanID:     price.PriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.PriceStdYear.UnitAmount,
				Edition:    price.PriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.PriceStdYear.UnitAmount,
					Currency: "cny",
				},
				Kind:          enum.OrderKindCreate,
				PaymentMethod: enum.PayMethodAli,
				WxAppID:       null.String{},
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
				DatePeriod:    dt.DatePeriod{},
				LiveMode:      true,
			},
			wantErr: false,
		},
		{
			name: "Renewal order",
			args: args{
				counter: subs.Counter{
					BaseAccount: newPersona.BaseAccount(),
					FtcPrice:    price.PriceStdYear,
					Method:      enum.PayMethodWx,
					WxAppID:     null.StringFrom(wxID),
				},
				p: renewalPerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    renewalPerson.AccountID(),
				PlanID:     price.PriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.PriceStdYear.UnitAmount,
				Edition:    price.PriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.PriceStdYear.UnitAmount,
					Currency: "cny",
				},
				Kind:          enum.OrderKindRenew,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
				DatePeriod:    dt.DatePeriod{},
				LiveMode:      true,
			},
			wantErr: false,
		},
		{
			name: "Upgrade order",
			args: args{
				counter: subs.Counter{
					BaseAccount: newPersona.BaseAccount(),
					FtcPrice:    price.PricePrm,
					Method:      enum.PayMethodWx,
					WxAppID:     null.StringFrom(wxID),
				},
				p: upgradePerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    upgradePerson.AccountID(),
				PlanID:     price.PricePrm.ID,
				DiscountID: null.String{},
				Price:      price.PricePrm.UnitAmount,
				Edition:    price.PricePrm.Edition,
				Charge: price.Charge{
					Amount:   price.PricePrm.UnitAmount,
					Currency: "cny",
				},
				Kind:          enum.OrderKindUpgrade,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
				DatePeriod:    dt.DatePeriod{},
				LiveMode:      true,
			},
			wantErr: false,
		},
		{
			name: "Add-on order",
			args: args{
				counter: subs.Counter{
					BaseAccount: newPersona.BaseAccount(),
					FtcPrice:    price.PriceStdYear,
					Method:      enum.PayMethodWx,
					WxAppID:     null.StringFrom(wxID),
				},
				p: addOnPerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    addOnPerson.AccountID(),
				PlanID:     price.PriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.PriceStdYear.UnitAmount,
				Edition:    price.PriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.PriceStdYear.UnitAmount,
					Currency: "cny",
				},
				Kind:          enum.OrderKindAddOn,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
				DatePeriod:    dt.DatePeriod{},
				LiveMode:      true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.args.p

			switch tt.want.Kind {
			case enum.OrderKindRenew:
				repo.MustSaveMembership(
					reader.NewMockMemberBuilder(p.FtcID).
						Build(),
				)
			case enum.OrderKindUpgrade:
				repo.MustSaveMembership(
					reader.NewMockMemberBuilder(p.FtcID).
						Build(),
				)
			case enum.OrderKindAddOn:
				repo.MustSaveMembership(
					reader.NewMockMemberBuilder(p.FtcID).
						WithPayMethod(enum.PayMethodStripe).
						Build(),
				)
			}
			got, err := env.CreateOrder(tt.args.counter)

			tt.want.ID = got.Order.ID
			tt.want.CreatedAt = got.Order.CreatedAt

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Order, tt.want) {
				t.Errorf("CreateOrder() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestEnv_LogOrderMeta(t *testing.T) {
	type fields struct {
		dbs    db.ReadWriteSplit
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
				dbs: test.SplitDB,
			},
			args: args{
				m: subs.OrderMeta{
					OrderID: pkg.MustOrderID(),
					Client:  faker.RandomClientApp(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
				logger: tt.fields.logger,
			}
			if err := env.SaveOrderMeta(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("SaveOrderMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveOrder(t *testing.T) {
	p := test.NewPersona()
	order := p.NewOrder(enum.OrderKindCreate)

	repo := test.NewRepo()
	repo.MustSaveOrder(order)

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve order",
			args: args{
				orderID: order.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

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
