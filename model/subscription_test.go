package model

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestEnv_CreateOrder(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	type args struct {
		user      paywall.UserID
		plan      paywall.Plan
		payMethod enum.PayMethod
		clientApp util.ClientApp
		wxAppId   null.String
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "New Member Order",
			args: args{
				user:      userID,
				plan:      test.YearlyStandard,
				payMethod: enum.PayMethodWx,
				clientApp: test.RandomClientApp(),
				wxAppId:   null.StringFrom(test.WxPayApp.AppID),
			},
			wantErr: false,
		},
		{
			name: "Renewal Order",
			args: args{
				user:      userID,
				plan:      test.YearlyStandard,
				payMethod: enum.PayMethodAli,
				clientApp: test.RandomClientApp(),
			},
			wantErr: false,
		},
		{
			name: "Upgrade Order",
			args: args{
				user:      userID,
				plan:      test.YearlyPremium,
				payMethod: enum.PayMethodAli,
				clientApp: test.RandomClientApp(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:    test.DB,
				query: query.NewBuilder(false),
			}

			got, err := env.CreateOrder(tt.args.user, tt.args.plan, tt.args.payMethod, tt.args.clientApp, tt.args.wxAppId)
			if err != nil {
				t.Errorf("Env.CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			confirmed, err := env.ConfirmPayment(got.ID, time.Now())

			if err != nil {
				t.Error(err)
			}

			t.Logf("Confirmed order %+v", confirmed)
		})
	}
}
