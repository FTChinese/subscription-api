package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
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

func TestEnv_VerifyPassword(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		params input.PasswordUpdateParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.AuthResult
		wantErr bool
	}{
		{
			name: "Verify password before permitting change",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				params: input.PasswordUpdateParams{
					FtcID: a.FtcID,
					Old:   a.Password,
					New:   "",
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
			got, err := env.VerifyPassword(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyPassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_UpdatePassword(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		p input.PasswordUpdateParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Update password",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				p: input.PasswordUpdateParams{
					FtcID: a.FtcID,
					Old:   "",
					New:   "23456789",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.UpdatePassword(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
