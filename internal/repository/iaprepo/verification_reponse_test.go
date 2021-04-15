package iaprepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

var cfg = config.NewBuildConfig(false, false)

func TestEnv_SaveDecodedReceipt(t *testing.T) {

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		v apple.VerifiedReceiptSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save receipt in verification response",
			args: args{
				v: test.MustVerificationResponse().ReceiptSchema(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveDecodedReceipt(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveDecodedReceipt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveTransaction(t *testing.T) {
	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		r apple.TransactionSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save a transaction",
			args: args{
				r: test.MustIAPTransaction().Schema(apple.EnvSandbox),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveTransaction(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SavePendingRenewal(t *testing.T) {
	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		p apple.PendingRenewalSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save a pending renewal",
			args: args{
				p: test.MustPendingRenewal().Schema(apple.EnvSandbox),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SavePendingRenewal(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SavePendingRenewal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
