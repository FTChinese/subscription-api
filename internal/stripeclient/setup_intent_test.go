package stripeclient

import (
	"github.com/FTChinese/subscription-api/faker"
	sdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_FetchSetupIntent(t *testing.T) {
	faker.MustSetupViper()

	c := New(false, zaptest.NewLogger(t))

	type args struct {
		id                  string
		expandPaymentMethod bool
	}
	tests := []struct {
		name    string
		args    args
		want    *sdk.SetupIntent
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id:                  "seti_1KUQKcBzTK0hABgJssUtMfl8",
				expandPaymentMethod: true,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.FetchSetupIntent(tt.args.id, tt.args.expandPaymentMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchSetupIntent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("FetchSetupIntent() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%+v", got)
		})
	}
}
