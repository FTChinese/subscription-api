package subrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_CreateOrder(t *testing.T) {
	p := test.NewPersona()
	repo := test.NewRepo()

	repo.MustSaveAccount(p.FtcAccount())

	type fields struct {
		db     *sqlx.DB
		cache  *cache.Cache
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
			name: "Alipay order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				config: subs.NewPayment(p.FtcAccount(), p.GetPlan()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				cache:  tt.fields.cache,
				logger: tt.fields.logger,
			}
			got, err := env.CreateOrder(tt.args.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("Order %+v", got.Order)
		})
	}
}

func TestEnv_ProratedOrdersUsed(t *testing.T) {
	upOrderID := subs.MustGenerateOrderID()

	pos := test.GenProratedOrders(upOrderID)

	test.NewRepo().
		MustSaveProratedOrders(pos)

	type fields struct {
		db     *sqlx.DB
		cache  *cache.Cache
		logger *zap.Logger
	}
	type args struct {
		upOrderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Flag prorated orders as used",
			fields: fields{
				db:     test.DB,
				cache:  nil,
				logger: nil,
			},
			args:    args{upOrderID: upOrderID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				cache:  tt.fields.cache,
				logger: tt.fields.logger,
			}
			if err := env.ProratedOrdersUsed(tt.args.upOrderID); (err != nil) != tt.wantErr {
				t.Errorf("ProratedOrdersUsed() error = %v, wantErr %v", err, tt.wantErr)
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
					OrderID: subs.MustGenerateOrderID(),
					Client:  faker.RandomClientApp(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				cache:  tt.fields.cache,
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
	order := p.CreateOrder()

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
				db:     tt.fields.db,
				cache:  tt.fields.cache,
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
	order := p.CreateOrder()

	t.Logf("Order id: %s", order.ID)

	test.NewRepo().MustSaveOrder(order)

	env := NewEnv(test.DB, test.Cache, zaptest.NewLogger(t))

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
