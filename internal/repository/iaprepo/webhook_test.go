package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"testing"
)

func TestIAPEnv_SaveNotification(t *testing.T) {

	wh := apple.WebHook{
		AutoRenewAdamID:              0,
		AutoRenewProductID:           "com.ft.ftchinese.mobile.subscription.member.monthly",
		AutoRenewStatus:              "0",
		AutoRenewStatusChangeDate:    "",
		AutoRenewStatusChangeDateMs:  "1532683981000",
		AutoRenewStatusChangeDatePST: "",
		Environment:                  apple.EnvSandbox,
		ExpirationIntent:             null.StringFrom("1"),
		LatestExpiredReceipt:         "",
		LatestExpiredReceiptInfo:     nil,
		LatestToken:                  "",
		LatestTransaction:            apple.LatestTransaction{},
		NotificationType:             apple.NotificationTypeInitialBuy,
		Password:                     "12345678",
		UnifiedReceipt: apple.UnifiedReceipt{
			Status: 0,
		},
	}

	env := Env{
		db: test.DB,
	}

	type args struct {
		w apple.WebHookSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save notification",
			args: args{
				w: wh.Schema(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWebhook(tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("SaveWebhook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateWebhook(t *testing.T) {
	wh := apple.WebHook{
		AutoRenewAdamID:              0,
		AutoRenewProductID:           "com.ft.ftchinese.mobile.subscription.member.monthly",
		AutoRenewStatus:              "0",
		AutoRenewStatusChangeDate:    "",
		AutoRenewStatusChangeDateMs:  "1532683981000",
		AutoRenewStatusChangeDatePST: "",
		Environment:                  apple.EnvSandbox,
		ExpirationIntent:             null.StringFrom("1"),
		LatestExpiredReceipt:         "",
		LatestExpiredReceiptInfo:     nil,
		LatestToken:                  "",
		LatestTransaction:            apple.LatestTransaction{},
		NotificationType:             apple.NotificationTypeInitialBuy,
		Password:                     "12345678",
		UnifiedReceipt: apple.UnifiedReceipt{
			Status: 0,
		},
	}

	b, err := json.Marshal(wh)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(b))
}
