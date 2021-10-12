package accounts

import (
	"github.com/FTChinese/go-rest/chrono"
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
					Mobile: v.Mobile,
					Code:   v.Code,
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

func TestEnv_SetMobile(t *testing.T) {

	noProfile := test.NewPersona()
	mobileTaken := test.NewPersona()
	mobileLinked := test.NewPersona()
	ftcWithoutMobile := test.NewPersona()
	ftcWithMobileTaken := test.NewPersona()

	env := New(test.SplitDB, zaptest.NewLogger(t))
	repo := test.NewRepo()

	type args struct {
		params account.MobileUpdater
	}
	tests := []struct {
		name        string
		baseAccount account.BaseAccount
		hasProfile  bool
		args        args
		wantErr     bool
	}{
		{
			name:        "Row does not exist",
			baseAccount: noProfile.EmailOnlyAccount(),
			hasProfile:  false,
			args: args{
				params: noProfile.MobileUpdater(),
			},
			wantErr: false,
		},
		{
			name:        "Mobile linked to another account",
			baseAccount: mobileTaken.EmailMobileAccount(),
			hasProfile:  true,
			args: args{
				params: test.NewPersona().
					WithMobile(mobileTaken.Mobile).
					MobileUpdater(),
			},
			wantErr: true,
		},
		{
			name:        "Mobile and ftc id linked",
			baseAccount: mobileLinked.EmailMobileAccount(),
			hasProfile:  true,
			args: args{
				params: mobileLinked.MobileUpdater(),
			},
			wantErr: true,
		},
		{
			name:        "FTC id exists without mobile",
			baseAccount: ftcWithoutMobile.EmailOnlyAccount(),
			hasProfile:  true,
			args: args{
				params: ftcWithoutMobile.MobileUpdater(),
			},
			wantErr: false,
		},
		{
			name: "FTC id exists with other mobile",
			baseAccount: ftcWithMobileTaken.
				EmailMobileAccount(),
			hasProfile: true,
			args: args{
				params: ftcWithMobileTaken.
					WithMobile(faker.GenPhone()).
					MobileUpdater(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.CreateUserInfo(tt.baseAccount)
			if tt.hasProfile {
				repo.CreateProfile(tt.baseAccount)
			}

			t.Logf("%v", tt.args.params)

			if err := env.UpsertMobile(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("UpsertMobile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
