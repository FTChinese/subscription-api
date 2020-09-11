package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
	"reflect"
	"testing"
)

func TestEnv_Link(t *testing.T) {
	type fields struct {
		cfg config.BuildConfig
		db  *sqlx.DB
	}
	type args struct {
		account reader.FtcAccount
		iapSubs apple.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    apple.LinkResult
		wantErr bool
	}{
		{
			name: "Existing IAP cannot link to another FTC account",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				cfg: tt.fields.cfg,
				db:  tt.fields.db,
			}
			got, err := env.Link(tt.args.account, tt.args.iapSubs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Link() got = %v, want %v", got, tt.want)
			}
		})
	}
}
