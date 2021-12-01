package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
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
		ExpirationIntent:             1,
		NotificationType:             apple.NotificationTypeInitialBuy,
		Password:                     "12345678",
		UnifiedReceipt: apple.UnifiedReceipt{
			Status: 0,
		},
	}

	env := Env{
		Env:    shared.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
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
				w: apple.NewWebHookSchema(wh),
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
