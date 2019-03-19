package model

import (
	"database/sql"
	"testing"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_findMember(t *testing.T) {
	m := newMocker().withUserID()
	mm := m.createMember()
	t.Logf("Created membership: %+v\n", mm)

	subs := m.wxpaySubs()

	type fields struct {
		sandbox bool
		db      *sql.DB
	}
	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Find Member",
			fields:  fields{db: db},
			args:    args{subs: subs},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.findMember(tt.args.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.findMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Got: %+v\n", got)
		})
	}
}
