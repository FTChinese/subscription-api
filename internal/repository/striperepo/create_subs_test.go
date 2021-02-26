package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type paymentAttached struct {
	account         reader.FtcAccount
	paymentMethodID string
}

func newCustomerAndPayment(client Client, account reader.FtcAccount) (paymentAttached, error) {
	defer client.logger.Sync()
	sugar := client.logger.Sugar()

	cus, err := client.CreateCustomer(account.Email)
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

	account.StripeID = null.StringFrom(cus.ID)

	return paymentAttached{
		account:         account,
		paymentMethodID: pm.ID,
	}, nil
}

func TestEnv_CreateSubscription(t *testing.T) {

	p := test.NewPersona()

	repo := test.NewRepo()
	m := p.Membership()
	t.Logf("%v", m.MemberID)

	repo.MustSaveMembership(m)

	pa, err := newCustomerAndPayment(
		NewClient(false, zaptest.NewLogger(t)),
		p.FtcAccount(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	type fields struct {
		db     *sqlx.DB
		client Client
		logger *zap.Logger
	}
	type args struct {
		params stripe.SubsParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create subscription",
			fields: fields{
				db:     test.DB,
				client: NewClient(false, zaptest.NewLogger(t)),
				logger: zaptest.NewLogger(t),
			},
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
			env := Env{
				db:     tt.fields.db,
				client: tt.fields.client,
				logger: tt.fields.logger,
			}

			got, err := env.CreateSubscription(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Subscription result: \n%s", faker.MustMarshalIndent(got))
		})
	}
}
