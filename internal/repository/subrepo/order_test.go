package subrepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
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
		counter ftcpay.Counter
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
				counter: ftcpay.Counter{
					BaseAccount: newPersona.EmailBaseAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: price.MockPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodAli,
					WxAppID:   null.String{},
				},
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    newPersona.AccountID(),
				PlanID:     price.MockPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.MockPriceStdYear.UnitAmount,
				Edition:    price.MockPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.MockPriceStdYear.UnitAmount,
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
				counter: ftcpay.Counter{
					BaseAccount: newPersona.EmailBaseAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: price.MockPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: renewalPerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    renewalPerson.AccountID(),
				PlanID:     price.MockPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.MockPriceStdYear.UnitAmount,
				Edition:    price.MockPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.MockPriceStdYear.UnitAmount,
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
				counter: ftcpay.Counter{
					BaseAccount: newPersona.EmailBaseAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: price.MockPricePrm.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: upgradePerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    upgradePerson.AccountID(),
				PlanID:     price.MockPricePrm.ID,
				DiscountID: null.String{},
				Price:      price.MockPricePrm.UnitAmount,
				Edition:    price.MockPricePrm.Edition,
				Charge: price.Charge{
					Amount:   price.MockPricePrm.UnitAmount,
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
				counter: ftcpay.Counter{
					BaseAccount: newPersona.EmailBaseAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: price.MockPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: addOnPerson,
			},
			want: subs.Order{
				ID:         "",
				UserIDs:    addOnPerson.AccountID(),
				PlanID:     price.MockPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      price.MockPriceStdYear.UnitAmount,
				Edition:    price.MockPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   price.MockPriceStdYear.UnitAmount,
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
		dbs    db.ReadWriteMyDBs
		logger *zap.Logger
	}
	type args struct {
		c footprint.OrderClient
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
				c: footprint.OrderClient{
					OrderID: ids.MustOrderID(),
					Client:  footprint.MockClient(""),
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
			if err := env.SaveOrderMeta(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("SaveOrderMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveOrder(t *testing.T) {

	ftcID := uuid.New().String()

	repo := test.NewRepo()
	order := repo.MustSaveOrder(subs.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())

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
	order := p.OrderBuilder().
		WithStdYear().
		WithCreate().
		WithAlipay().
		Build()

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

func TestEnv_ListOrders(t *testing.T) {
	ftcID := uuid.New().String()

	repo := test.NewRepo()
	repo.MustSaveOrder(subs.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(subs.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(subs.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())

	type fields struct {
		Env    readers.Env
		logger *zap.Logger
	}
	type args struct {
		ids ids.UserIDs
		p   gorest.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "List orders",
			fields: fields{
				Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
				logger: zaptest.NewLogger(t),
			},
			args: args{
				ids: ids.UserIDs{
					CompoundID: "",
					FtcID:      null.StringFrom(ftcID),
					UnionID:    null.String{},
				}.MustNormalize(),
				p: gorest.NewPagination(1, 10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env:    tt.fields.Env,
				logger: tt.fields.logger,
			}
			got, err := env.ListOrders(tt.args.ids, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
