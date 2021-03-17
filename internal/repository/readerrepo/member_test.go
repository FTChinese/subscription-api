package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_RetrieveMember(t *testing.T) {
	p := test.NewPersona()
	m := p.Membership()
	test.NewRepo().MustSaveMembership(m)

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		id pkg.UserIDs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load member",
			args: args{
				id: m.UserIDs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})
	}
}
