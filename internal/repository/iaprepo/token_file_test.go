package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveReceiptTokenFile(t *testing.T) {
	type args struct {
		r apple.ReceiptSchema
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
			if err := SaveReceiptToDisk(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptToDisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReceipt(t *testing.T) {
	type args struct {
		s apple.BaseSchema
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
				s: apple.BaseSchema{
					OriginalTransactionID: rt.OriginalTransactionID,
					Environment:           rt.Environment,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadReceiptFromDisk(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReceiptFromDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotNil(t, got)

			t.Logf("Token: %s", got)
		})
	}
}
