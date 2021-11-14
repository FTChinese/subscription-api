package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
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
		p.EmailOnlyAccount())
	if err != nil {
		t.Error(err)
		return
	}

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		client: NewClient(false, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	_, err = env.CreateSubscription(
		pa.account,
		stripe.CheckoutItem{
			Price:        stripe.Price{},
			Introductory: stripe.Price{},
		},
		stripe.SubSharedParams{
			DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		ba     account.BaseAccount
		item   stripe.CheckoutItem
		params stripe.SubSharedParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Switch cycle",
			args: args{
				ba: pa.account,
				item: stripe.CheckoutItem{
					Price:        stripe.Price{},
					Introductory: stripe.Price{},
				},
				params: stripe.SubSharedParams{
					DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
				},
			},
			wantErr: false,
		},
		{
			name: "Upgrade",
			args: args{
				ba: pa.account,
				item: stripe.CheckoutItem{
					Price:        stripe.Price{},
					Introductory: stripe.Price{},
				},
				params: stripe.SubSharedParams{
					DefaultPaymentMethod: null.StringFrom(pa.paymentMethodID),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.UpdateSubscription(
				tt.args.ba, tt.args.item, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
