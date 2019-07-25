package model

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_PreviewUpgrade(t *testing.T) {

	userID := test.NewProfile().RandomUserID()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	for i := 0; i < 2; i++ {
		subs, err := env.CreateOrder(
			userID,
			test.YearlyStandard,
			enum.PayMethodAli,
			test.RandomClientApp(),
			null.String{})
		if err != nil {
			panic(err)
		}

		t.Logf("Created an order: %+v", subs)

		subs, err = env.ConfirmPayment(subs.ID, time.Now())
		if err != nil {
			t.Error(err)
			panic(err)
		}

		t.Logf("Confirmed an order: %+v", subs)
	}

	type args struct {
		userID paywall.AccountID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Preview Upgrade",
			args: args{
				userID: userID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.PreviewUpgrade(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.PreviewUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Upgrade preview: %+v", got)
		})
	}
}
