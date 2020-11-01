package readerrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestEnv_ArchiveMember(t *testing.T) {
	p := test.NewPersona()

	ss := p.Membership().Snapshot(reader.FtcArchiver(enum.OrderKindRenew))

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		snapshot reader.MemberSnapshot
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Archiving old membership",
			fields: fields{db: test.DB},
			args: args{
				snapshot: ss,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			if err := env.ArchiveMember(tt.args.snapshot); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
