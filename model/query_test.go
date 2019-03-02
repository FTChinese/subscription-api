package model

import (
	"database/sql"
	"testing"

	"github.com/patrickmn/go-cache"
)

func TestEnv_stmtInsertSubs(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Sandbox Insert Subscription Statement",
			fields: fields{sandbox: true},
		},
		{
			name:   "Production Insert Subscription Statement",
			fields: fields{sandbox: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtInsertSubs()
			t.Logf("%s\n", got)
		})
	}
}

func TestEnv_stmtSelectSubs(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Sandbox Select Subscription",
			fields: fields{sandbox: true},
		},
		{
			name:   "Production Select Subscription",
			fields: fields{sandbox: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtSelectSubs()
			t.Logf("%s\n", got)
		})
	}
}

func TestEnv_stmtSelectSubsLock(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Sandbox Select Subscription with Row Locking",
			fields: fields{sandbox: true},
		},
		{
			name:   "Production Select Subscrption with Row Locking",
			fields: fields{sandbox: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtSelectSubsLock()
			t.Logf("%s\n", got)
		})
	}
}

func TestEnv_stmtUpdateSubs(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Sandbox Update Subscription Statement",
			fields: fields{sandbox: true},
		},
		{
			name:   "Production Update Subscription Statement",
			fields: fields{sandbox: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtUpdateSubs()
			t.Logf("%s", got)
		})
	}
}
