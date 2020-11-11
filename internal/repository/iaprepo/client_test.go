package iaprepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_Verify(t *testing.T) {
	type args struct {
		receipt string
		sandbox bool
	}
	tests := []struct {
		name    string
		fields  Client
		args    args
		wantErr bool
	}{
		{
			name:   "A sandbox receipt",
			fields: NewClient(zaptest.NewLogger(t)),
			args: args{
				receipt: faker.IAPReceipt,
				sandbox: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			got, err := c.Verify(tt.args.receipt, tt.args.sandbox)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Verified receipt: %v", got)
		})
	}
}
