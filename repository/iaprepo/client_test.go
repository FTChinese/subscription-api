package iaprepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"testing"
)

func TestClient_pickUrl(t *testing.T) {
	config.MustSetupViper()

	tests := []struct {
		name   string
		fields Client
		want   string
	}{
		{
			name:   "Sandbox",
			fields: NewClient(true),
			want:   "https://sandbox.itunes.apple.com/verifyReceipt",
		},
		{
			name:   "Production",
			fields: NewClient(false),
			want:   "https://buy.itunes.apple.com/verifyReceipt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			if got := c.pickUrl(); got != tt.want {
				t.Errorf("pickUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Verify(t *testing.T) {
	type args struct {
		receipt string
	}
	tests := []struct {
		name    string
		fields  Client
		args    args
		wantErr bool
	}{
		{
			name:    "A sandbox receipt",
			fields:  NewClient(true),
			args:    args{receipt: faker.IAPReceipt},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				isSandbox:  tt.fields.isSandbox,
				sandboxUrl: tt.fields.sandboxUrl,
				prodUrl:    tt.fields.prodUrl,
				password:   tt.fields.password,
			}
			got, err := c.Verify(tt.args.receipt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Verified receipt: %s", got)
		})
	}
}
