package model

import (
	"database/sql"
	"testing"

	"github.com/FTChinese/go-rest"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

func TestEnv_SaveWxAccess(t *testing.T) {
	m := newMocker()
	acc := m.wxAccess()
	t.Logf("OAuth Access: %+v\n", acc)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		appID string
		acc   wxlogin.OAuthAccess
		c     gorest.ClientApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Save Wecaht OAuth Access",
			fields: fields{db: db},
			args: args{
				appID: oauthApp.AppID,
				acc:   acc,
				c:     clientApp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.SaveWxAccess(tt.args.appID, tt.args.acc, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_LoadWxAccess(t *testing.T) {
	m := newMocker()
	acc := m.createWxAccess()

	t.Logf("Created access token: %+v\n", acc)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		appID     string
		sessionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "Load Wechat OAuth Access",
			fields: fields{db: db},
			args: args{
				appID:     oauthApp.AppID,
				sessionID: acc.SessionID,
			},
			want: acc.SessionID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
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
	m := newMocker()
	acc := m.createWxAccess()

	t.Logf("Original access: %+v\n", acc)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		sessionID   string
		accessToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Update Wechat OAuth Access",
			fields: fields{db: db},
			args: args{
				sessionID:   acc.SessionID,
				accessToken: generateToken(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.UpdateWxAccess(tt.args.sessionID, tt.args.accessToken); (err != nil) != tt.wantErr {
				t.Errorf("Env.UpdateWxAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxUser(t *testing.T) {
	m := newMocker().withUnionID()
	userInfo := m.wxUserInfo()

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		u wxlogin.UserInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Save Wechat User Info",
			fields: fields{db: db},
			args: args{
				u: userInfo,
			},
			wantErr: false,
		},
		{
			name:   "Save Existing Wechat User",
			fields: fields{db: db},
			args: args{
				u: m.wxUserInfo(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.SaveWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_UpdateWxUser(t *testing.T) {
	m := newMocker().withUnionID()
	userInfo := m.createWxUser()
	t.Logf("Created wechat user: %+v\n", userInfo)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		u wxlogin.UserInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Update Wechat User Info",
			fields: fields{db: db},
			args: args{
				u: m.wxUserInfo(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.UpdateWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("Env.UpdateWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxStatus(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		code    int64
		message string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Save Wechat OAuth Response Status",
			fields: fields{db: db},
			args: args{
				code:    40029,
				message: "invalid code",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.SaveWxStatus(tt.args.code, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Generate multiple OAuthAccess with same OpenID,
// and multiple UserInfo with same OpenID and UnionID.
func TestMultiLogin(t *testing.T) {
	m := newMocker()

	for i := 0; i < 5; i++ {
		acc := m.createWxAccess()
		t.Logf("A new login: %+v\n", acc)

		userInfo := m.createWxUser()

		t.Logf("Save/Update userinfo: %+v\n", userInfo)
	}
}
