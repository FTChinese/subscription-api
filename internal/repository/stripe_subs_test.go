package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_RetrieveSubs(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	subs := test.NewPersona().StripeSubsBuilder().Build()

	test.NewRepo().SaveStripeSubs(subs)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.Subs
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: subs.ID,
			},
			want:    stripe.Subs{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrieveSubs(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveSubs() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}

func TestStripeRepo_UpsertSubs(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		s        stripe.Subs
		expanded bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				s:        test.NewPersona().StripeSubsBuilder().Build(),
				expanded: true,
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				s:        test.NewPersona().StripeSubsBuilder().Build(),
				expanded: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpsertSubs(tt.args.s, tt.args.expanded); (err != nil) != tt.wantErr {
				t.Errorf("UpsertSubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
