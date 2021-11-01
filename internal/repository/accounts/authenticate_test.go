package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestEnv_Authenticate(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		params input.EmailLoginParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.AuthResult
		wantErr bool
	}{
		{
			name: "Authenticate email and password",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				params: input.EmailLoginParams{
					EmailCredentials: input.EmailCredentials{
						Email:    a.Email,
						Password: a.Password,
					},
					DeviceToken: null.StringFrom(uuid.New().String()),
				},
			},
			want: account.AuthResult{
				UserID:          a.FtcID,
				PasswordMatched: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.Authenticate(tt.args.params.EmailCredentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Authenticate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_SignUpCount(t *testing.T) {
	faker.SeedGoFake()

	ip := gofakeit.IPv4Address()

	t.Logf("IP %s", ip)

	test.NewRepo().MustSaveFootprintN(footprint.
		NewMockFootprintBuilder(ip).
		WithSource(footprint.SourceSignUp).
		BuildN(5))

	type fields struct {
		Env readers.Env
	}
	type args struct {
		params account.SignUpRateParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.SignUpLimit
		wantErr bool
	}{
		{
			name: "Count sign-up of same ip",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				params: account.NewSignUpRateParams(ip, 1),
			},
			want: account.SignUpLimit{
				Count: 5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.SignUpCount(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignUpCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SignUpCount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_VerifyIDPassword(t *testing.T) {
	p := test.NewPersona()
	a := p.EmailOnlyAccount()

	repo := test.NewRepo()
	repo.CreateFtcAccount(a)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		params account.IDCredentials
	}
	tests := []struct {
		name    string
		args    args
		want    account.AuthResult
		wantErr bool
	}{
		{
			name: "Verify id and password",
			args: args{
				params: account.IDCredentials{
					FtcID:    a.FtcID,
					Password: a.Password,
				},
			},
			want: account.AuthResult{
				UserID:          a.FtcID,
				PasswordMatched: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.VerifyIDPassword(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyIDPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyIDPassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_UpdatePassword(t *testing.T) {
	p := test.NewPersona()
	a := p.EmailOnlyAccount()

	repo := test.NewRepo()
	repo.CreateFtcAccount(a)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		p account.IDCredentials
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update password",
			args: args{
				p: account.IDCredentials{
					FtcID:    a.FtcID,
					Password: "0987654321",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdatePassword(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			result, err := env.VerifyIDPassword(tt.args.p)
			if err != nil {
				t.Errorf("VerifyIDPassowrd() error = %v", err)
				return
			}

			if !result.PasswordMatched {
				t.Errorf("Changing password failed")
			}
		})
	}
}
