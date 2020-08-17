package products

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ftchinese/subscription-api/pkg/product"
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

func TestEnv_PlanByID(t *testing.T) {
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
			got, err := env.PlanByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrievePlanByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestEnv_PlanByEdition(t *testing.T) {
	type fields struct {
		db    *sqlx.DB
		cache *cache.Cache
	}
	type args struct {
		e product.Edition
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Load a plan by tier and cycle",
			fields: fields{
				db:    test.DB,
				cache: test.Cache,
			},
			args: args{
				e: product.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
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
			got, err := env.PlanByEdition(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("PlanByEdition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}
