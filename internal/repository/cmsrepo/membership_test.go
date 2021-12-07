package cmsrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_CreateMembership(t *testing.T) {
	p := test.NewPersona()

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		ba     account.BaseAccount
		params input.MemberParams
	}
	tests := []struct {
		name    string
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Create membership",
			args: args{
				ba: p.EmailOnlyAccount(),
				params: input.MemberParams{
					Tier:       enum.TierStandard,
					Cycle:      enum.CycleYear,
					ExpireDate: chrono.DateFrom(time.Now().AddDate(0, 6, 0)),
					PayMethod:  enum.PayMethodAli,
					PriceID:    price.MockPriceStdYear.ID,
				},
			},
			want:    reader.Membership{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.CreateMembership(tt.args.ba, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMembership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("CreateMembership() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_UpdateMembership(t *testing.T) {
	p := test.NewPersona()

	m := p.MemberBuilder().Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		compoundID string
		params     input.MemberParams
		by         string
	}
	tests := []struct {
		name    string
		args    args
		want    reader.MembershipVersioned
		wantErr bool
	}{
		{
			name: "Update membership",
			args: args{
				compoundID: m.CompoundID,
				params: input.MemberParams{
					Tier:       enum.TierStandard,
					Cycle:      enum.CycleYear,
					ExpireDate: chrono.DateFrom(time.Now().AddDate(0, 0, 7)),
					PayMethod:  enum.PayMethodWx,
					PriceID:    price.MockPriceStdYear.ID,
				},
				by: "test",
			},
			want:    reader.MembershipVersioned{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.UpdateMembership(tt.args.compoundID, tt.args.params, tt.args.by)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMembership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("UpdateMembership() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

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
