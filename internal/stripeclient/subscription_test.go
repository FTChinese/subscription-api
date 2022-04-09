package stripeclient

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_NewSubs(t *testing.T) {
	faker.MustSetupViper()

	client := New(false, zaptest.NewLogger(t))

	type args struct {
		params *stripe.SubscriptionParams
	}
	tests := []struct {
		name    string
		fields  Client
		args    args
		want    *stripe.Subscription
		wantErr bool
	}{
		{
			name:    "Introductory offer",
			fields:  client,
			args:    args{},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				sc:     tt.fields.sc,
				logger: tt.fields.logger,
			}
			got, err := c.NewSubs(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("NewSubs() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
