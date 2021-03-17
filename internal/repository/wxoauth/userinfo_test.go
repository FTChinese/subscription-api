package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_SaveWxUser(t *testing.T) {
	env := Env{
		dbs: test.SplitDB,
	}

	type args struct {
		u wxlogin.UserInfoSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save wx user",
			args: args{
				u: test.NewPersona().WxUser(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("SaveWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
