package iaprepo

import (
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestIAPEnv_VerifyReceipt(t *testing.T) {
	resp := test.GetVerificationResponse()
	resp.Parse()

	env := IAPEnv{
		c:  util.NewBuildConfig(false, false),
		db: nil,
	}

	type args struct {
		r apple.VerificationRequestBody
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Verify Receipt",
			args: args{
				r: apple.VerificationRequestBody{ReceiptData: test.GetReceiptToken()},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.VerifyReceipt(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Verification response: %+v", got)
		})
	}
}

func TestIAPEnv_SaveVerificationSession(t *testing.T) {

	resp := test.GetVerificationResponse()
	resp.Parse()

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		v apple.VerificationSessionSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Verification Session",
			args: args{
				v: resp.SessionSchema(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveVerificationSession(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveVerificationSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_SaveTransaction(t *testing.T) {
	resp := test.GetVerificationResponse()
	resp.Parse()

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
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
			name: "Save Transaction",
			args: args{
				r: test.GetIAPTransaction().Schema(apple.EnvSandbox),
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

func TestIAPEnv_SavePendingRenewal(t *testing.T) {

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
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
			name: "Save Pending Renewal",
			args: args{
				p: test.GetPendingRenewal().Schema(apple.EnvSandbox),
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

func TestIAPEnv_CreateSubscription(t *testing.T) {

	resp := test.GetVerificationResponse()
	resp.Parse()

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		s apple.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create Subscription",
			args: args{
				s: resp.Subscription(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.CreateSubscription(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("CreateSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_SaveReceiptToken(t *testing.T) {
	resp := test.GetVerificationResponse()
	resp.Parse()

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}
	type args struct {
		r apple.ReceiptToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Receipt Token",
			args: args{
				r: resp.ReceiptToken(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveReceiptToken(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
