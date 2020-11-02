package subrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		result subs.PaymentResult
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Confirm ali order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				result: p1.PaymentResult(order1),
			},
			wantErr: false,
		},
		{
			name: "Confirm wx order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				result: p1.PaymentResult(order2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}

			t.Logf("Payment result: %+v", tt.args.result)

			got, err := env.ConfirmOrder(tt.args.result)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
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
				db:     tt.fields.db,
				cache:  tt.fields.cache,
				logger: tt.fields.logger,
			}
			if err := env.SaveConfirmationErr(tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("SaveConfirmationErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
