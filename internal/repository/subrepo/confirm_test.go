package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	repo := test.NewRepo()

	p1 := test.NewPersona()
	aliCreateOrder := p1.NewOrder(enum.OrderKindCreate)
	t.Logf("Ali Order id %s", aliCreateOrder.ID)

	p2 := test.NewPersona().SetPayMethod(enum.PayMethodWx)
	wxCreateOrder := p2.NewOrder(enum.OrderKindCreate)
	t.Logf("Wx Order id %s", wxCreateOrder.ID)

	p3 := test.NewPersona().SetAccountKind(enum.AccountKindLinked)
	linkedAccountOrder := p3.NewOrder(enum.OrderKindCreate)
	t.Logf("Order for linked account %s", linkedAccountOrder.ID)

	// Order confirmed but not synced to membership
	p4 := test.NewPersona()
	outOfSyncOrder := subs.NewMockOrderBuilder("").
		WithUserIDs(p4.AccountID()).
		WithPrice(faker.PriceStdYear).
		WithKind(enum.OrderKindRenew).
		WithPayMethod(enum.PayMethodAli).
		WithConfirmed().
		WithStartTime(time.Now()).
		Build()
	t.Logf("Out of sync order %v", outOfSyncOrder)

	p5 := test.NewPersona()
	memberPriorRenewal := p5.Membership()
	renewalOrder := p5.NewOrder(enum.OrderKindRenew)

	p6 := test.NewPersona()
	memberPriorUpgrade := p6.Membership()
	upgradeOrder := p6.NewOrder(enum.OrderKindUpgrade)

	p7 := test.NewPersona().SetPayMethod(enum.PayMethodApple)
	iapMember := p7.Membership()
	p7.SetPayMethod(enum.PayMethodAli)
	addOnOrder := p7.NewOrder(enum.OrderKindAddOn)

	env := NewEnv(test.DB, zaptest.NewLogger(t))

	type args struct {
		result subs.PaymentResult
		order  subs.Order
	}
	type requisite struct {
		currentMember reader.Membership
	}
	tests := []struct {
		name      string
		requisite requisite
		args      args
		wantErr   bool
	}{
		{
			name: "confirm new ali order",
			args: args{
				result: subs.MockNewPaymentResult(aliCreateOrder),
				order:  aliCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm new wx order",
			args: args{
				result: subs.MockNewPaymentResult(wxCreateOrder),
				order:  wxCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed new linked account order",
			args: args{
				result: subs.MockNewPaymentResult(linkedAccountOrder),
				order:  linkedAccountOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed out of sync order",
			args: args{
				result: subs.MockNewPaymentResult(outOfSyncOrder),
				order:  outOfSyncOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm renewal",
			requisite: requisite{
				currentMember: memberPriorRenewal,
			},
			args: args{
				result: subs.MockNewPaymentResult(renewalOrder),
				order:  renewalOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm upgrade",
			requisite: requisite{
				currentMember: memberPriorUpgrade,
			},
			args: args{
				result: subs.MockNewPaymentResult(upgradeOrder),
				order:  upgradeOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm add-on",
			requisite: requisite{
				currentMember: iapMember,
			},
			args: args{
				result: subs.MockNewPaymentResult(addOnOrder),
				order:  addOnOrder,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prerequisite.
			if !tt.requisite.currentMember.IsZero() {
				repo.MustSaveMembership(tt.requisite.currentMember)
			}
			repo.MustSaveOrder(tt.args.order)

			got, err := env.ConfirmOrder(tt.args.result, tt.args.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_SaveConfirmationErr(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		cache  *cache.Cache
		logger *zap.Logger
	}
	type args struct {
		e *subs.ConfirmError
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Confirmation error",
			fields: fields{
				db: test.DB,
			},
			args: args{
				e: &subs.ConfirmError{
					OrderID: db.MustOrderID(),
					Message: "Test error",
					Retry:   false,
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
			if err := env.SaveConfirmErr(tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("SaveConfirmErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SavePayResult(t *testing.T) {

	env := NewEnv(test.DB, zaptest.NewLogger(t))

	type args struct {
		result subs.PaymentResult
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save payment result",
			args: args{
				result: subs.PaymentResult{
					PaymentState:     ali.TradeStatusSuccess,
					PaymentStateDesc: "",
					Amount:           null.IntFrom(28000),
					TransactionID:    faker.GenTxID(),
					OrderID:          db.MustOrderID(),
					PaidAt:           chrono.TimeNow(),
					PayMethod:        enum.PayMethodAli,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SavePayResult(tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("SavePayResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
