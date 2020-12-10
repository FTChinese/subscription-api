package striperepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func MustNewInvoice() stripe.Invoice {
	var in stripeSdk.Invoice
	if err := json.Unmarshal([]byte(faker.StripeInvoice), &in); err != nil {
		panic(err)
	}

	return stripe.NewInvoice(&in)
}

func TestEnv_UpsertInvoice(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		client Client
		logger *zap.Logger
	}
	type args struct {
		i stripe.Invoice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Upsert invoice",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				i: MustNewInvoice(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := env.UpsertInvoice(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("UpsertInvoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
