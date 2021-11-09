package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_WxSignUp(t *testing.T) {
	faker.SeedGoFake()

	p := test.NewPersona()
	w := p.WxUser()

	test.NewRepo().MustSaveWxUser(w)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		unionID string
		input   input.EmailSignUpParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.WxEmailLinkResult
		wantErr bool
	}{
		{
			name: "Wx signup",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				unionID: w.UnionID,
				input: input.EmailSignUpParams{
					EmailCredentials: input.EmailCredentials{
						Email:    gofakeit.Email(),
						Password: "12345678",
					},
					DeviceToken: null.StringFrom(uuid.New().String()),
					SourceURL:   "",
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
			got, err := env.WxSignUp(tt.args.unionID, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WxSignUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("WxSignUp() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_LinkWechat(t *testing.T) {
	faker.SeedGoFake()

	p1 := test.NewPersona()
	p2 := test.NewPersona()
	ftcA := p1.EmailOnlyAccount()
	wxA := p1.EmailWxAccount()

	w := p2.WxUser()

	repo := test.NewRepo()
	repo.MustSaveWxUser(w)
	repo.MustCreateFtcAccount(ftcA)

	linked, err := reader.WxEmailLinkBuilder{
		FTC: reader.Account{
			BaseAccount: ftcA,
			LoginMethod: enum.LoginMethodEmail,
			Wechat:      account.Wechat{},
			Membership:  reader.Membership{},
		},
		Wechat: reader.Account{
			BaseAccount: wxA,
			LoginMethod: enum.LoginMethodWx,
			Wechat: account.Wechat{
				WxNickname:  w.NickName,
				WxAvatarURL: w.AvatarURL,
			},
			Membership: reader.Membership{},
		},
	}.Build()

	if err != nil {
		t.Error(err)
		return
	}

	type fields struct {
		Env readers.Env
	}
	type args struct {
		result reader.WxEmailLinkResult
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.WxEmailLinkResult
		wantErr bool
	}{
		{
			name: "Link wechat",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				result: linked,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			err := env.LinkWechat(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("WxLinkEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("WxLinkEmail() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestEnv_UnlinkWx(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		acnt   reader.Account
		anchor enum.AccountKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Unlink wechat",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				acnt: reader.Account{
					BaseAccount: a,
					LoginMethod: enum.LoginMethodEmail,
					Wechat:      account.Wechat{},
					Membership:  reader.Membership{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.UnlinkWx(tt.args.acnt, tt.args.anchor); (err != nil) != tt.wantErr {
				t.Errorf("WxUnlinkEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
