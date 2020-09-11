package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveReceiptTokenFile(t *testing.T) {
	type args struct {
		r apple.ReceiptToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save receipt",
			args: args{
				r: test.MustVerificationResponse().ReceiptToken(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveReceiptTokenFile(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptTokenFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReceipt(t *testing.T) {
	type args struct {
		originalID string
		env        apple.Environment
	}
	rt := test.MustVerificationResponse().ReceiptToken()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load Receipt",
			args: args{
				originalID: rt.OriginalTransactionID,
				env:        rt.Environment,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadReceipt(tt.args.originalID, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotNil(t, got)

			t.Logf("Token: %s", got)
		})
	}
}
