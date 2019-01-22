package model

import (
	"database/sql"
	"testing"

	cache "github.com/patrickmn/go-cache"
	"github.com/smartwalle/alipay"
)

func TestEnv_SaveAliNotification(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		n alipay.TradeNotification
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Save Ali Notification",
			fields:  fields{db: db},
			args:    args{n: aliNoti()},
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
			if err := env.SaveAliNotification(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveAliNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
