package products

import (
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_loadBanner(t *testing.T) {
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
			name: "Load banner",
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
			got, err := env.loadBanner()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
			assert.NotEmpty(t, got.Banner.ID)
		})
	}
}

func TestEnv_loadProducts(t *testing.T) {
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
			name: "Load products for paywall",
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
			got, err := env.loadProducts()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)

			assert.Len(t, got, 2)
		})
	}
}

func TestEnv_retrievePaywall(t *testing.T) {
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
			name: "Load paywall",
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
			got, err := env.retrievePaywall()
			if (err != nil) != tt.wantErr {
				t.Errorf("retrievePaywall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}
