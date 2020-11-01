package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestEnv_SaveWxUser(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		u wxlogin.UserInfoSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save wx user",
			fields: fields{
				db: test.DB,
			},
			args: args{
				u: test.NewPersona().WxUser(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			if err := env.SaveWxUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("SaveWxUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
