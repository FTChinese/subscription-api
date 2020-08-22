package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

var cfg = config.NewBuildConfig(false, false)

func TestEnv_SaveVerifiedReceipt(t *testing.T) {
	type fields struct {
		cfg config.BuildConfig
		db  *sqlx.DB
	}
	type args struct {
		v apple.VerifiedReceiptSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save receipt in verification response",
			fields: fields{
				cfg: cfg,
				db:  test.DB,
			},
			args: args{
				v: test.MustVerificationResponse().ReceiptSchema(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				cfg: tt.fields.cfg,
				db:  tt.fields.db,
			}
			if err := env.SaveVerifiedReceipt(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveVerifiedReceipt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
