package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"reflect"
	"testing"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	type fields struct {
		BuildConfig config.BuildConfig
		db          *sqlx.DB
		cache       *cache.Cache
	}
	type args struct {
		result subs.PaymentResult
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   subs.ConfirmationResult
		want1  *subs.ConfirmError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				BuildConfig: tt.fields.BuildConfig,
				db:          tt.fields.db,
				cache:       tt.fields.cache,
			}
			got, got1 := env.ConfirmOrder(tt.args.result)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfirmOrder() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ConfirmOrder() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
