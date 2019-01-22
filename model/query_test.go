package model

import (
	"database/sql"
	"testing"

	cache "github.com/patrickmn/go-cache"
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

func TestEnv_stmtInsertMember(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Sandbox Insert Member Statement",
			fields: fields{sandbox: true},
		},
		{
			name:   "Production Insert Member Statement",
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
			got := env.stmtInsertMember()
			t.Logf("%s\n", got)
		})
	}
}

func TestEnv_stmtSelectMember(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		isWxLogin bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Sandbox Select Member Statement for Email Login",
			fields: fields{sandbox: true},
			args:   args{isWxLogin: false},
		},
		{
			name:   "Sandbox Select Member Statement for Wechat Login",
			fields: fields{sandbox: true},
			args:   args{isWxLogin: true},
		},
		{
			name:   "Production Select Member Statement for Email Login",
			fields: fields{sandbox: false},
			args:   args{isWxLogin: false},
		},
		{
			name:   "Production Select Member Statement for Wechat Login",
			fields: fields{sandbox: false},
			args:   args{isWxLogin: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtSelectMember(tt.args.isWxLogin)
			t.Logf("%s\n", got)
		})
	}
}

func TestEnv_stmtSelectExpLock(t *testing.T) {
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		isWxLogin bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Sandbox Expiration Date with Row Locking and Email Login",
			fields: fields{sandbox: true},
			args:   args{isWxLogin: false},
		},
		{
			name:   "Sandbox Expiration Date with Row Locking and Wechat Login",
			fields: fields{sandbox: true},
			args:   args{isWxLogin: true},
		},
		{
			name:   "Production Expiration Date with Row Locking and Email Login",
			fields: fields{sandbox: false},
			args:   args{isWxLogin: false},
		},
		{
			name:   "Sandbox Expiration Date with Row Locking and Wechat Login",
			fields: fields{sandbox: false},
			args:   args{isWxLogin: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got := env.stmtSelectExpLock(tt.args.isWxLogin)

			t.Logf("%s\n", got)
		})
	}
}
