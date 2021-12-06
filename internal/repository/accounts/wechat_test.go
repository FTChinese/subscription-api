package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_WxSignUp(t *testing.T) {
	faker.SeedGoFake()

	p := test.NewPersona()
	w := p.WxUser()

	test.NewRepo().MustSaveWxUser(w)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		merged reader.Account
	}
	tests := []struct {
		name    string
		args    args
		want    reader.WxEmailLinkResult
		wantErr bool
	}{
		{
			name: "Wx signup",
			args: args{
				merged: reader.Account{
					BaseAccount: p.EmailWxAccount(),
					LoginMethod: enum.LoginMethodWx,
					Wechat:      p.Wechat(),
					Membership:  p.MemberBuilder().Build(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.WxSignUp(reader.Account{
				BaseAccount: account.BaseAccount{},
				LoginMethod: 0,
				Wechat:      account.Wechat{},
				Membership:  reader.Membership{},
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("WxSignUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("WxSignUp() got = %v, want %v", got, tt.want)
			//}
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

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		result reader.WxEmailLinkResult
	}
	tests := []struct {
		name    string
		args    args
		want    reader.WxEmailLinkResult
		wantErr bool
	}{
		{
			name: "Link wechat",
			args: args{
				result: linked,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		acnt   reader.Account
		anchor enum.AccountKind
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Unlink wechat",
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
			if err := env.UnlinkWx(tt.args.acnt, tt.args.anchor); (err != nil) != tt.wantErr {
				t.Errorf("WxUnlinkEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
