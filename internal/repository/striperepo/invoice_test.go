package striperepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	stripeSdk "github.com/stripe/stripe-go/v72"
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
	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		client: Client{},
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		i stripe.Invoice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Upsert invoice",
			args: args{
				i: MustNewInvoice(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.UpsertInvoice(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("UpsertInvoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
