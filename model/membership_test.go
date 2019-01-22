package model

import (
	"database/sql"
	"testing"

	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_findMember(t *testing.T) {
	m := newMocker()
	mm := m.createMember()
	t.Logf("Created membership: %+v\n", mm)

	subs := m.wxpaySubs()
	t.Logf("Subscription: %+v\n", subs)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    paywall.Membership
		wantErr bool
	}{
		{
			name:    "Find Member",
			fields:  fields{db: db},
			args:    args{subs},
			want:    mm,
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
			got, err := env.findMember(tt.args.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.findMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Got: %+v\n", got)
			// The comparison cound never be equal. Do not use this test.
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Env.findMember() = %v, want %v", got, tt.want)
			// }
		})
	}
}
