package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

type paymentAttached struct {
	account         account.BaseAccount
	paymentMethodID string
}

func newCustomerAndPayment(client stripeclient.Client, acnt account.BaseAccount) (paymentAttached, error) {

	cus, err := client.CreateCustomer(acnt.Email)
	if err != nil {
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
		return paymentAttached{}, err
	}

	_, err = client.AttachPaymentMethod(cus.ID, pm.ID)
	if err != nil {
		return paymentAttached{}, err
	}

	acnt.StripeID = null.StringFrom(cus.ID)

	return paymentAttached{
		account:         acnt,
		paymentMethodID: pm.ID,
	}, nil
}

func TestEnv_CreateSubscription(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t), shared.StripeBaseRepo{
		Client: stripeclient.New(false, zaptest.NewLogger(t)),
		Live:   false,
		Cache:  nil,
	})

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
			name: "Create subscription",
			args: args{
				ba: account.BaseAccount{
					FtcID:        "c07f79dc-664b-44ca-87ea-42958e7991b0",
					UnionID:      null.String{},
					StripeID:     null.StringFrom("cus_IXp31Fk2jYJmU3"),
					Email:        "stripe.test@ftchinese.com",
					Mobile:       null.String{},
					UserName:     null.String{},
					AvatarURL:    null.String{},
					IsVerified:   false,
					CampaignCode: null.String{},
				},
				item: stripe.CheckoutItem{
					Price:        stripe.MockPriceStdYear,
					Introductory: stripe.MockPriceStdIntro,
				},
				params: stripe.SubSharedParams{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.CreateSubscription(tt.args.ba, tt.args.item, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Subscription result: \n%s", faker.MustMarshalIndent(got))
		})
	}
}
