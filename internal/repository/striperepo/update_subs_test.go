package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_UpdateSubscription(t *testing.T) {
	p := test.NewPersona()

	client := NewClient(false, zaptest.NewLogger(t))

	pa, err := newCustomerAndPayment(
		client,
		p.FtcAccount())
	if err != nil {
		t.Error(err)
		return
	}

	env := Env{
		dbs:    test.SplitDB,
		client: NewClient(false, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	_, err = env.CreateSubscription(stripe.SubsParams{
		Account: pa.account,
		Edition: price.StripeEditions.MustFindByEdition(price.StdMonthEdition, false),
		SharedParams: stripe.SharedParams{
			DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		cfg stripe.SubsParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Switch cycle",
			args: args{
				cfg: stripe.SubsParams{
					Account: pa.account,
					Edition: price.StripeEditions.MustFindByEdition(price.StdYearEdition, false),
					SharedParams: stripe.SharedParams{
						DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Upgrade",
			args: args{
				cfg: stripe.SubsParams{
					Account: pa.account,
					Edition: price.StripeEditions.MustFindByEdition(price.PremiumEdition, false),
					SharedParams: stripe.SharedParams{
						DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.UpdateSubscription(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
