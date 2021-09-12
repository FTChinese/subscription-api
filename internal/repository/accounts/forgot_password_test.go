package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SavePwResetSession(t *testing.T) {
	faker.SeedGoFake()

	type fields struct {
		Env readers.Env
	}
	type args struct {
		s account.PwResetSession
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save password reset session",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				s: account.MustNewPwResetSession(input.ForgotPasswordParams{
					Email:     gofakeit.Email(),
					SourceURL: null.String{},
				}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.SavePwResetSession(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SavePwResetSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_PwResetSessionByToken(t *testing.T) {
	faker.SeedGoFake()

	sess := account.MustNewPwResetSession(input.ForgotPasswordParams{
		Email:     gofakeit.Email(),
		SourceURL: null.String{},
	})

	_ = New(test.SplitDB, zaptest.NewLogger(t)).SavePwResetSession(sess)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.PwResetSession
		wantErr bool
	}{
		{
			name: "Retrieve password reset session",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				token: sess.URLToken,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.PwResetSessionByToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("PwResetSessionByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("PwResetSessionByToken() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_PwResetSessionByCode(t *testing.T) {
	faker.SeedGoFake()

	sess := account.MustNewPwResetSession(input.ForgotPasswordParams{
		Email:     gofakeit.Email(),
		SourceURL: null.String{},
	}).WithPlatform(enum.PlatformAndroid)

	_ = New(test.SplitDB, zaptest.NewLogger(t)).SavePwResetSession(sess)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		params input.AppResetPwSessionParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.PwResetSession
		wantErr bool
	}{
		{
			name: "Retrieve password reset session for mobile app",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				params: input.AppResetPwSessionParams{
					Email:   sess.Email,
					AppCode: sess.AppCode.String,
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
			got, err := env.PwResetSessionByCode(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("PwResetSessionByCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("PwResetSessionByCode() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_DisablePasswordReset(t *testing.T) {
	faker.SeedGoFake()

	sess := account.MustNewPwResetSession(input.ForgotPasswordParams{
		Email:     gofakeit.Email(),
		SourceURL: null.String{},
	})

	_ = New(test.SplitDB, zaptest.NewLogger(t)).SavePwResetSession(sess)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Disable password reset session",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				t: sess.URLToken,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.DisablePasswordReset(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("DisablePasswordReset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
