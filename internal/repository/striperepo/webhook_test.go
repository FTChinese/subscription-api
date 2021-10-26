package striperepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
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

func TestEnv_OnWebhookSubs(t *testing.T) {

	p := test.NewPersona()
	ftcValid := test.NewPersona()

	repo := test.NewRepo()
	repo.MustSaveMembership(ftcValid.MemberBuilder().Build())

	env := New(db.MockMySQL(), NewClient(false, zaptest.NewLogger(t)), zaptest.NewLogger(t))

	type args struct {
		subs    stripe.Subs
		userIDs ids.UserIDs
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.WebhookResult
		wantErr bool
	}{
		{
			name: "Create stripe membership",
			args: args{
				subs:    stripe.NewMockSubsBuilder(p.FtcID).Build(),
				userIDs: p.UserIDs(),
			},
			want:    stripe.WebhookResult{},
			wantErr: false,
		},
		{
			name: "Valid ftc membership overridden and carried over",
			args: args{
				subs:    stripe.NewMockSubsBuilder(ftcValid.FtcID).Build(),
				userIDs: ftcValid.UserIDs(),
			},
			want:    stripe.WebhookResult{},
			wantErr: false,
		},
		{
			name: "Expired stripe cannot override valid ftc membership",
			args: args{
				subs:    stripe.NewMockSubsBuilder(ftcValid.FtcID).WithCanceled().Build(),
				userIDs: ftcValid.UserIDs(),
			},
			want:    stripe.WebhookResult{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.OnWebhookSubs(tt.args.subs, tt.args.userIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnWebhookSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("OnWebhookSubs() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
