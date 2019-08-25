package repository

import (
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/test"
	"reflect"
	"testing"
)

func TestEnv_AddMemberID(t *testing.T) {

	profile := test.NewProfile()

	m := profile.Membership(reader.AccountKindFtc)
	m.ID = null.String{}

	test.NewRepo().SaveMember(m)

	t.Logf("Saved member %+v", m)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		m paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Add Member ID If Missing",
			args: args{
				m: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.m.GenerateID()

			if err := env.AddMemberID(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("AddMemberID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_BackUpMember(t *testing.T) {
	m := test.NewProfile().Membership(reader.AccountKindFtc)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		m paywall.MemberSnapshot
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Backup Membership",
			args: args{
				m: paywall.NewMemberSnapshot(m),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.BackUpMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("BackUpMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_FindUnusedOrders(t *testing.T) {
	type fields struct {
		BuildConfig BuildConfig
		db          *sqlx.DB
		cache       *cache.Cache
		query       query.Builder
	}
	type args struct {
		id reader.AccountID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []paywall.ProrationSource
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				BuildConfig: tt.fields.BuildConfig,
				db:          tt.fields.db,
				cache:       tt.fields.cache,
				query:       tt.fields.query,
			}
			got, err := env.FindUnusedOrders(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindUnusedOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindUnusedOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_RetrieveMember(t *testing.T) {
	type fields struct {
		BuildConfig BuildConfig
		db          *sqlx.DB
		cache       *cache.Cache
		query       query.Builder
	}
	type args struct {
		id reader.AccountID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    paywall.Membership
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				BuildConfig: tt.fields.BuildConfig,
				db:          tt.fields.db,
				cache:       tt.fields.cache,
				query:       tt.fields.query,
			}
			got, err := env.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveMember() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_RetrieveUpgradePlan(t *testing.T) {
	type fields struct {
		BuildConfig BuildConfig
		db          *sqlx.DB
		cache       *cache.Cache
		query       query.Builder
	}
	type args struct {
		upgradeID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    paywall.UpgradePlan
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				BuildConfig: tt.fields.BuildConfig,
				db:          tt.fields.db,
				cache:       tt.fields.cache,
				query:       tt.fields.query,
			}
			got, err := env.RetrieveUpgradePlan(tt.args.upgradeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveUpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveUpgradePlan() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_RetrieveUpgradeSource(t *testing.T) {
	type fields struct {
		BuildConfig BuildConfig
		db          *sqlx.DB
		cache       *cache.Cache
		query       query.Builder
	}
	type args struct {
		upgradeID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []paywall.ProrationSource
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				BuildConfig: tt.fields.BuildConfig,
				db:          tt.fields.db,
				cache:       tt.fields.cache,
				query:       tt.fields.query,
			}
			got, err := env.RetrieveUpgradeSource(tt.args.upgradeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveUpgradeSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveUpgradeSource() got = %v, want %v", got, tt.want)
			}
		})
	}
}
