package products

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestEnv_loadPlans(t *testing.T) {
	type fields struct {
		db    *sqlx.DB
		cache *cache.Cache
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "List paywall plans",
			fields: fields{
				db:    test.DB,
				cache: test.Cache,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:    tt.fields.db,
				cache: tt.fields.cache,
			}
			got, err := env.loadPlans()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadPlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)

			assert.Len(t, got, 3)
		})
	}
}

func TestEnv_retrievePlan(t *testing.T) {
	type fields struct {
		db    *sqlx.DB
		cache *cache.Cache
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Load a single plan",
			fields: fields{
				db:    test.DB,
				cache: test.Cache,
			},
			args:    args{id: "plan_2cc3ncDcKiM7"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:    tt.fields.db,
				cache: tt.fields.cache,
			}
			got, err := env.retrievePlan(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrievePlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}
