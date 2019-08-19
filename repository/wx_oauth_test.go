package repository

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
	"testing"

	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

func TestEnv_SaveWxAccess(t *testing.T) {

	env := Env{
		db: test.DB,
	}

	type args struct {
		appID string
		acc   wxlogin.OAuthAccess
		c     util.ClientApp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Wechat OAuth Access",
			args: args{
				appID: test.WxOAuthApp.AppID,
				acc:   test.MyProfile.WxAccess(),
				c:     test.RandomClientApp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWxAccess(tt.args.appID, tt.args.acc, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_LoadWxAccess(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	acc := test.MyProfile.WxAccess()

	err := env.SaveWxAccess(test.WxOAuthApp.AppID, acc, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	type args struct {
		appID     string
		sessionID string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Load Wechat OAuth Access",
			args: args{
				appID:     test.WxOAuthApp.AppID,
				sessionID: acc.SessionID,
			},
			want: acc.SessionID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.LoadWxAccess(tt.args.appID, tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.LoadWxAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.SessionID != acc.SessionID {
				t.Errorf("Env.LoadWxAccess() expected %s, got %s", acc.SessionID, got.SessionID)
			}
		})
	}
}

func TestEnv_UpdateWxAccess(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	acc := test.MyProfile.WxAccess()
	if err := env.SaveWxAccess(test.WxOAuthApp.AppID, acc, test.RandomClientApp()); err != nil {
		panic(err)
	}

	t.Logf("Original access: %+v\n", acc)

	type args struct {
		sessionID   string
		accessToken string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update Wechat OAuth Access",
			args: args{
				sessionID:   acc.SessionID,
				accessToken: test.GenToken(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateWxAccess(tt.args.sessionID, tt.args.accessToken); (err != nil) != tt.wantErr {
				t.Errorf("Env.UpdateWxAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxUserEmoji(t *testing.T) {
	u := wxlogin.UserInfo{
		UnionID:    test.MyProfile.UnionID,
		NickName:   "😈🤠",
		AvatarURL:  test.GenAvatar(),
		Sex:        0,
		Country:    "China",
		Province:   "Beijing",
		City:       "Beijing",
		Privileges: []string{},
	}

	t.Logf("Emoji: %+v", u)
}

func TestEnv_SaveWxUser(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	type args struct {
		u wxlogin.UserInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Wechat AccountID Info",
			args: args{
				u: test.MyProfile.WxInfo(),
			},
			wantErr: false,
		},
		{
			name: "Save Emoji",
			args: args{
				u: wxlogin.UserInfo{
					UnionID:    test.MyProfile.UnionID,
					NickName:   "😈🤠",
					AvatarURL:  test.GenAvatar(),
					Sex:        0,
					Country:    "China",
					Province:   "Beijing",
					City:       "Beijing",
					Privileges: []string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_UpdateWxUser(t *testing.T) {

	env := Env{
		db: test.DB,
	}

	type args struct {
		u wxlogin.UserInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update Wechat AccountID Info",
			args: args{
				u: test.MyProfile.WxInfo(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.UpdateWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("Env.UpdateWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxStatus(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	type args struct {
		code    int64
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Wechat OAuth Response Status",
			args: args{
				code:    40029,
				message: "invalid code",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWxStatus(tt.args.code, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
