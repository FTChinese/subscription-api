package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
