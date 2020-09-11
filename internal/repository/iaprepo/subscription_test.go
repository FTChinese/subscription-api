package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestEnv_LoadSubscription(t *testing.T) {
	type fields struct {
		cfg config.BuildConfig
		db  *sqlx.DB
	}
	type args struct {
		originalID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Load a subscription",
			fields: fields{
				cfg: test.CFG,
				db:  test.DB,
			},
			args:    args{originalID: "1000000619244062"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				cfg: tt.fields.cfg,
				db:  tt.fields.db,
			}
			got, err := env.LoadSubscription(tt.args.originalID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", got.Environment)
		})
	}
}
