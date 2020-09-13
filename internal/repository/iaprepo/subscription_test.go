package iaprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_LoadSubscription(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db *sqlx.DB
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
				db: test.DB,
			},
			args:    args{originalID: p.AppleSubID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.LoadSubs(tt.args.originalID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			t.Logf("%s", got.Environment)
		})
	}
}

func TestEnv_countSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db *sqlx.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Count iap subscription",
			fields: fields{
				db: test.DB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.countSubs()
			if (err != nil) != tt.wantErr {
				t.Errorf("countSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("Total rows %d", got)
		})
	}
}

func TestEnv_listSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		p gorest.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "List subs",
			fields: fields{
				db: test.DB,
			},
			args: args{
				p: gorest.NewPagination(1, 20),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.listSubs(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("listSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ListSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		p gorest.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Count and list subs",
			fields: fields{
				db: test.DB,
			},
			args: args{
				p: gorest.NewPagination(1, 20),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.ListSubs(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
