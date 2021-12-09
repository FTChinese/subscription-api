package subrepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	subs2 "github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
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

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		counter subs2.Counter
		p       *test.Persona
	}
	tests := []struct {
		name    string
		args    args
		want    subs2.Order
		wantErr bool
	}{
		{
			name: "New order",
			args: args{
				counter: subs2.Counter{
					BaseAccount: newPersona.EmailOnlyAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: pw.MockPwPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodAli,
					WxAppID:   null.String{},
				},
			},
			want: subs2.Order{
				ID:         "",
				UserIDs:    newPersona.UserIDs(),
				PlanID:     pw.MockPwPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      pw.MockPwPriceStdYear.UnitAmount,
				Edition:    pw.MockPwPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   pw.MockPwPriceStdYear.UnitAmount,
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
				counter: subs2.Counter{
					BaseAccount: newPersona.EmailOnlyAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: pw.MockPwPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: renewalPerson,
			},
			want: subs2.Order{
				ID:         "",
				UserIDs:    renewalPerson.UserIDs(),
				PlanID:     pw.MockPwPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      pw.MockPwPriceStdYear.UnitAmount,
				Edition:    pw.MockPwPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   pw.MockPwPriceStdYear.UnitAmount,
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
				counter: subs2.Counter{
					BaseAccount: newPersona.EmailOnlyAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: pw.MockPwPricePrm.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: upgradePerson,
			},
			want: subs2.Order{
				ID:         "",
				UserIDs:    upgradePerson.UserIDs(),
				PlanID:     pw.MockPwPricePrm.ID,
				DiscountID: null.String{},
				Price:      pw.MockPwPricePrm.UnitAmount,
				Edition:    pw.MockPwPricePrm.Edition,
				Charge: price.Charge{
					Amount:   pw.MockPwPricePrm.UnitAmount,
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
				counter: subs2.Counter{
					BaseAccount: newPersona.EmailOnlyAccount(),
					CheckoutItem: price.CheckoutItem{
						Price: pw.MockPwPriceStdYear.Price,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: addOnPerson,
			},
			want: subs2.Order{
				ID:         "",
				UserIDs:    addOnPerson.UserIDs(),
				PlanID:     pw.MockPwPriceStdYear.ID,
				DiscountID: null.String{},
				Price:      pw.MockPwPriceStdYear.UnitAmount,
				Edition:    pw.MockPwPriceStdYear.Edition,
				Charge: price.Charge{
					Amount:   pw.MockPwPriceStdYear.UnitAmount,
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

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		c footprint.OrderClient
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Log order meta data",
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
			if err := env.SaveOrderMeta(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("SaveOrderMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveOrder(t *testing.T) {

	ftcID := uuid.New().String()

	repo := test.NewRepo()
	order := repo.MustSaveOrder(subs2.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

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

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

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
	repo.MustSaveOrder(subs2.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(subs2.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(subs2.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		ids ids.UserIDs
		p   gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "List orders",
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
			got, err := env.ListOrders(tt.args.ids, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
