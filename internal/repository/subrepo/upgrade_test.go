package subrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_UpgradeIntent(t *testing.T) {
	p := test.NewPersona()

	repo := test.NewRepo()

	orders := p.CreateBalanceSources(3)
	repo.MustSaveRenewalOrders(orders)
	repo.MustSaveMembership(p.Member)

	t.Logf("%v", p.Member)

	pFree := test.NewPersona()
	orders = pFree.CreateBalanceSources(10)
	repo.MustSaveRenewalOrders(orders)
	repo.MustSaveMembership(pFree.Member)

	t.Logf("%v", pFree.Member)

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
		wantErr bool
	}{
		{
			name: "Upgrade Preview",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.NewPayment(p.FtcAccount(), faker.PlanPrm).WithPreview(true),
			},
			wantErr: false,
		},
		{
			name: "Free upgrade",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				config: subs.NewPayment(pFree.FtcAccount(), faker.PlanPrm),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := env.UpgradeIntent(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeIntent() error = %v, wantErr %v", err, tt.wantErr)
				return
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
