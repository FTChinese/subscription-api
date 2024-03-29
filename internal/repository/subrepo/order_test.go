package subrepo

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestEnv_CreateOrder(t *testing.T) {
	wxID := faker.WxUnionID()
	repo := test.NewRepo()

	newPersona := test.NewPersona()
	renewalPerson := test.NewPersona()
	upgradePerson := test.NewPersona()
	addOnPerson := test.NewPersona()

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		counter reader.ShoppingCart
		p       *test.Persona
	}
	tests := []struct {
		name    string
		args    args
		want    ftcpay.Order
		wantErr bool
	}{
		{
			name: "New order",
			args: args{
				counter: reader.ShoppingCart{
					Account: newPersona.EmailOnlyAccount(),
					FtcItem: reader.CartItemFtc{
						Price: reader.MockPwPriceStdYear.FtcPrice,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodAli,
					WxAppID:   null.String{},
				},
			},
			want: ftcpay.Order{
				ID:            "",
				UserIDs:       newPersona.UserIDs(),
				Tier:          reader.MockPwPriceStdYear.Tier,
				Cycle:         reader.MockPwPriceStdYear.Cycle,
				Kind:          enum.OrderKindCreate,
				OriginalPrice: reader.MockPwPriceStdYear.UnitAmount,
				PayableAmount: reader.MockPwPriceStdYear.UnitAmount,
				PaymentMethod: enum.PayMethodAli,
				YearsCount:    0,
				MonthsCount:   0,
				DaysCount:     0,
				WxAppID:       null.String{},
				ConfirmedAt:   chrono.Time{},
				CreatedAt:     chrono.Time{},
				StartDate:     chrono.Date{},
				EndDate:       chrono.Date{},
			},
			wantErr: false,
		},
		{
			name: "Renewal order",
			args: args{
				counter: reader.ShoppingCart{
					Account: newPersona.EmailOnlyAccount(),
					FtcItem: reader.CartItemFtc{
						Price: reader.MockPwPriceStdYear.FtcPrice,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: renewalPerson,
			},
			want: ftcpay.Order{
				ID:            "",
				UserIDs:       renewalPerson.UserIDs(),
				OriginalPrice: reader.MockPwPriceStdYear.UnitAmount,
				Tier:          enum.TierStandard,
				Cycle:         enum.CycleYear,
				PayableAmount: reader.MockPwPriceStdYear.UnitAmount,
				Kind:          enum.OrderKindRenew,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
			},
			wantErr: false,
		},
		{
			name: "Upgrade order",
			args: args{
				counter: reader.ShoppingCart{
					Account: newPersona.EmailOnlyAccount(),
					FtcItem: reader.CartItemFtc{
						Price: reader.MockPwPricePrm.FtcPrice,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: upgradePerson,
			},
			want: ftcpay.Order{
				ID:            "",
				UserIDs:       upgradePerson.UserIDs(),
				OriginalPrice: reader.MockPwPricePrm.UnitAmount,
				Tier:          enum.TierPremium,
				Cycle:         enum.CycleYear,
				PayableAmount: reader.MockPwPricePrm.UnitAmount,
				Kind:          enum.OrderKindUpgrade,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
			},
			wantErr: false,
		},
		{
			name: "Add-on order",
			args: args{
				counter: reader.ShoppingCart{
					Account: newPersona.EmailOnlyAccount(),
					FtcItem: reader.CartItemFtc{
						Price: reader.MockPwPriceStdYear.FtcPrice,
						Offer: price.Discount{},
					},
					PayMethod: enum.PayMethodWx,
					WxAppID:   null.StringFrom(wxID),
				},
				p: addOnPerson,
			},
			want: ftcpay.Order{
				ID:            "",
				UserIDs:       addOnPerson.UserIDs(),
				OriginalPrice: reader.MockPwPriceStdYear.UnitAmount,
				Tier:          enum.TierStandard,
				Cycle:         enum.CycleYear,
				PayableAmount: reader.MockPwPriceStdYear.UnitAmount,
				Kind:          enum.OrderKindAddOn,
				PaymentMethod: enum.PayMethodWx,
				WxAppID:       null.StringFrom(wxID),
				CreatedAt:     chrono.Time{},
				ConfirmedAt:   chrono.Time{},
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
					reader.NewMockMemberBuilder().
						SetFtcID(p.FtcID).
						Build(),
				)
			case enum.OrderKindUpgrade:
				repo.MustSaveMembership(
					reader.NewMockMemberBuilder().
						SetFtcID(p.FtcID).
						Build(),
				)
			case enum.OrderKindAddOn:
				repo.MustSaveMembership(
					reader.NewMockMemberBuilder().
						SetFtcID(p.FtcID).
						WithStripe("").
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
	order := repo.MustSaveOrder(ftcpay.NewMockOrderBuilder("").
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

		})
	}
}

func TestEnv_ListOrders(t *testing.T) {
	ftcID := uuid.New().String()

	repo := test.NewRepo()
	repo.MustSaveOrder(ftcpay.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(ftcpay.NewMockOrderBuilder("").
		WithFtcID(ftcID).
		WithKind(enum.OrderKindCreate).
		Build())
	repo.MustSaveOrder(ftcpay.NewMockOrderBuilder("").
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
