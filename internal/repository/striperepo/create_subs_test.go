package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

type paymentAttached struct {
	account         account.BaseAccount
	paymentMethodID string
}

func newCustomerAndPayment(client Client, acnt account.BaseAccount) (paymentAttached, error) {
	defer client.logger.Sync()
	sugar := client.logger.Sugar()

	cus, err := client.CreateCustomer(acnt.Email)
	if err != nil {
		sugar.Error(err)
		return paymentAttached{}, err
	}

	pm, err := client.NewPaymentMethod(&stripeSdk.PaymentMethodCardParams{
		CVC:      stripeSdk.String("123"),
		ExpMonth: stripeSdk.String("02"),
		ExpYear:  stripeSdk.String("22"),
		Number:   stripeSdk.String("4242424242424242"),
		Token:    nil,
	})
	if err != nil {
		sugar.Error(err)
		return paymentAttached{}, err
	}

	si, err := client.AttachPaymentMethod(cus.ID, pm.ID)
	if err != nil {
		sugar.Error(err)
		return paymentAttached{}, err
	}

	sugar.Infof("%v", si)

	acnt.StripeID = null.StringFrom(cus.ID)

	return paymentAttached{
		account:         acnt,
		paymentMethodID: pm.ID,
	}, nil
}

func TestEnv_CreateSubscription(t *testing.T) {

	p := test.NewPersona()
	m := reader.NewMockMemberBuilder(p.FtcID).Build()
	t.Logf("%v", m.UserIDs)

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	pa, err := newCustomerAndPayment(
		NewClient(false, zaptest.NewLogger(t)),
		p.EmailBaseAccount(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	env := New(db.MockMySQL(), NewClient(false, zaptest.NewLogger(t)), zaptest.NewLogger(t))

	type args struct {
		params stripe.SubsParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create subscription",
			args: args{
				params: stripe.SubsParams{
					Account: pa.account,
					Edition: price.StripeEditions.MustFindByEdition(price.StdYearEdition, false),
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

			got, err := env.CreateSubscription(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Subscription result: \n%s", faker.MustMarshalIndent(got))
		})
	}
}
