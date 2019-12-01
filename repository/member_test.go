package repository

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
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
		m subscription.Membership
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
		m subscription.MemberSnapshot
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Backup Membership",
			args: args{
				m: subscription.NewMemberSnapshot(m, subscription.SubsKindRenew),
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

func TestEnv_RetrieveMember(t *testing.T) {

	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)

	store.MustRenewN(test.YearlyStandardLive, 1)

	test.NewRepo().SaveMember(store.Member)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}
	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Retrieve membership",
			args:    args{id: store.AccountID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Membership: %+v", got)
		})
	}
}

func TestEnv_FindBalanceSources(t *testing.T) {
	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)
	orders := store.MustRenewN(test.YearlyStandardLive, 3)

	testRepo := test.NewRepo()
	for _, v := range orders {
		testRepo.SaveOrder(v)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Find Balance Sources",
			args: args{
				id: store.AccountID,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindBalanceSources(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindBalanceSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("FindBalanceSources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_RetrieveUpgradePlan(t *testing.T) {

	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)

	store.MustRenewN(test.YearlyStandardLive, 3)
	store.MustCreate(test.YearlyPremiumLive)

	test.NewRepo().SaveUpgradePlan(store.UpgradePlan)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		upgradeID string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Retrieve Upgrade Plan",
			args: args{
				upgradeID: store.UpgradePlan.ID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveUpgradePlan(tt.args.upgradeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveUpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Upgrade plan: %+v", got)
		})
	}
}

func TestEnv_RetrieveProratedOrders(t *testing.T) {

	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)

	store.MustRenewN(test.YearlyStandardLive, 3)
	store.MustCreate(test.YearlyPremiumLive)

	test.NewRepo().SaveBalanceSources(store.UpgradePlan.Data)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		upgradeID string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Retrieve Upgrade ",
			args:    args{upgradeID: store.UpgradePlan.ID},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveProratedOrders(tt.args.upgradeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveProratedOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("RetrieveProratedOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}
