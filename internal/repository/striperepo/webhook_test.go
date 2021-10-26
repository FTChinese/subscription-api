package striperepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SaveWebhookError(t *testing.T) {
	env := New(db.MockMySQL(), NewClient(false, zaptest.NewLogger(t)), zaptest.NewLogger(t))

	p := test.NewPersona()

	type args struct {
		whe stripe.WebhookError
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Insert webhook error",
			args: args{
				whe: stripe.WebhookError{
					ID:                      faker.GenStripeSubID(),
					EventType:               "customer.subscription.created",
					Message:                 "cannot override valid non-stripe membership",
					CurrentStripeMembership: reader.MembershipJSON{},
					CurrentDestMembership: reader.MembershipJSON{
						Membership: p.MemberBuilder().Build(),
					},
					TargetUserID: p.FtcID,
					CreatedUTC:   chrono.TimeNow(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveWebhookError(tt.args.whe); (err != nil) != tt.wantErr {
				t.Errorf("SaveWebhookError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
