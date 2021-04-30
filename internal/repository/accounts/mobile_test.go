package accounts

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
	"time"
)

func TestEnv_SaveSMSVerifier(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	faker.SeedGoFake()

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	type args struct {
		v ztsms.Verifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save mobile verifier",
			args: args{
				v: v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveSMSVerifier(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveSMSVerifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveSMSVerifier(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	faker.SeedGoFake()

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	v.CreatedUTC = chrono.TimeFrom(time.Now().In(time.UTC).Truncate(time.Second))

	_ = env.SaveSMSVerifier(v)

	type args struct {
		params ztsms.VerifierParams
	}
	tests := []struct {
		name    string
		args    args
		want    ztsms.Verifier
		wantErr bool
	}{
		{
			name: "Retrieve sms verifier",
			args: args{
				params: ztsms.VerifierParams{
					Mobile:      v.Mobile,
					Code:        v.Code,
					DeviceToken: "",
				},
			},
			want:    v,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RetrieveSMSVerifier(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveSMSVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveSMSVerifier() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_SMSVerifierUsed(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	_ = env.SaveSMSVerifier(v)

	type args struct {
		v ztsms.Verifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag sms code as used",
			args: args{
				v: v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.v = tt.args.v.WithUsed()

			if err := env.SMSVerifierUsed(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SMSVerifierUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SetPhone(t *testing.T) {

	env := New(test.SplitDB, zaptest.NewLogger(t))

	acnt := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(acnt)

	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set mobile phone",
			args: args{
				a: acnt,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SetPhone(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("SetPhone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
