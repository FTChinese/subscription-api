package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	repo := test.NewRepo()

	p1 := test.NewPersona()
	order1 := p1.CreateOrder()
	t.Logf("Ali Order id %s", order1.ID)
	repo.MustSaveOrder(order1)

	p2 := test.NewPersona()
	order2 := p2.CreateOrder()
	t.Logf("Wx Order id %s", order2.ID)
	repo.MustSaveOrder(order2)

	p3 := test.NewPersona().SetAccountKind(enum.AccountKindWx)
	repo.MustSaveMembership(p3.Membership())
	p3.SetAccountKind(enum.AccountKindLinked)
	order3 := p3.CreateOrder()
	t.Logf("Order for linked account %s", order3.ID)
	repo.MustSaveOrder(order3)

	env := NewEnv(test.DB, zaptest.NewLogger(t))

	type args struct {
		result subs.PaymentResult
		order  subs.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirm ali order",
			args: args{
				result: p1.PaymentResult(order1),
				order:  order1,
			},
			wantErr: false,
		},
		{
			name: "Confirm wx order",
			args: args{
				result: p1.PaymentResult(order2),
				order:  order2,
			},
			wantErr: false,
		},
		{
			name: "Confirmed linked account order",
			args: args{
				result: p3.PaymentResult(order3),
				order:  order3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("Payment result: %+v", tt.args.result)

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
					OrderID: subs.MustGenerateOrderID(),
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
					OrderID:          subs.MustGenerateOrderID(),
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
