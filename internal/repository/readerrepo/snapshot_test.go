package readerrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_ArchiveMember(t *testing.T) {
	p := test.NewPersona()

	ss := p.Membership().Snapshot(reader.FtcArchiver(enum.OrderKindRenew))

	env := Env{
		dbs:    test.SplitDB,
		logger: nil,
	}
	type args struct {
		snapshot reader.MemberSnapshot
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archiving old membership",
			args: args{
				snapshot: ss,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.ArchiveMember(tt.args.snapshot); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
