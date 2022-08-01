package cmsrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_DeleteMembership(t *testing.T) {
	p := test.NewPersona()

	m := p.MemberBuilder().Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		compoundID string
	}
	tests := []struct {
		name    string
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Delete membership",
			args: args{
				compoundID: m.CompoundID,
			},
			want:    reader.Membership{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.DeleteMembership(tt.args.compoundID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMembership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("DeleteMembership() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
