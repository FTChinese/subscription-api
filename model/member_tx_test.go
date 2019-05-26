package model

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
)

func TestMemberTx_RetrieveOrder(t *testing.T) {
	u := test.MyProfile.RandomUser()
	subs1 := test.MyProfile.BuildSubs(u, enum.PayMethodWx, paywall.SubsKindRenew)
	subs2 := test.MyProfile.SubsUpgrade(u)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	err := env.SaveSubscription(subs1, test.RandomClientApp())
	if err != nil {
		panic(err)
	}
	err = env.SaveSubscription(subs2, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve a Renew Order",
			args: args{
				orderID: subs1.OrderID,
			},
			wantErr: false,
		},
		{
			name: "Retrieve Upgrade Order",
			args: args{
				orderID: subs2.OrderID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := m.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("MemberTx.RetireveOrder() = %+v", got)
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}

func TestMemberTx_RetrieveMember(t *testing.T) {
	p := test.NewProfile()
	u := p.RandomUser()
	subs1 := test.NewProfile().SubsRandom(u)
	subs2 := test.NewProfile().SubsRandom(u)

	// Insert a member to test an existing case.
	subs2, err := subs2.ConfirmWithMember(paywall.Membership{}, time.Now())
	if err != nil {
		panic(err)
	}

	mm, err := subs2.BuildMembership()
	if err != nil {
		panic(err)
	}

	err = test.NewModel(test.DB).CreateMember(mm)
	if err != nil {
		panic(err)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		args    args
		want    paywall.Membership
		wantErr bool
	}{
		{
			name: "Membership Does Not Exist",
			args: args{
				subs: subs1,
			},
			wantErr: false,
		},
		{
			name: "Membership already exist",
			args: args{
				subs: subs2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.RetrieveMember(tt.args.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("MemberTx.RetrieveMember() = %v", got)
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}

func TestMemberTx_InvalidUpgrade(t *testing.T) {
	u := test.MyProfile.RandomUser()
	subs1 := test.NewProfile().SubsUpgrade(u)
	subs2 := test.NewProfile().SubsUpgrade(u)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	err := env.SaveSubscription(subs1, test.RandomClientApp())
	if err != nil {
		panic(err)
	}
	err = env.SaveSubscription(subs2, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		orderID    string
		errInvalid error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Duplicate upgrade",
			args: args{
				orderID:    subs1.OrderID,
				errInvalid: paywall.ErrDuplicateUpgrading,
			},
			wantErr: false,
		},
		{
			name: "Member does not exist",
			args: args{
				orderID:    subs2.OrderID,
				errInvalid: paywall.ErrNoUpgradingTarget,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.InvalidUpgrade(tt.args.orderID, tt.args.errInvalid)

			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.InvalidUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}

func TestMemberTx_ConfirmOrder(t *testing.T) {
	p := test.NewProfile()
	u := p.RandomUser()
	subs := test.NewProfile().SubsRandom(u)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	err := env.SaveSubscription(subs, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	subs, err = subs.ConfirmWithMember(paywall.Membership{}, time.Now())
	if err != nil {
		panic(err)
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirm order",
			args: args{
				subs: subs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := m.ConfirmOrder(tt.args.subs); (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}

func TestMemberTx_MarkOrdersProrated(t *testing.T) {
	p := test.NewProfile()
	u := p.RandomUser()
	subs1 := p.SubsRandom(u)
	subs2 := p.SubsRandom(u)
	upSubs := p.SubsUpgrade(u)
	upSubs.UpgradeSource = []string{subs1.OrderID, subs2.OrderID}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	err := env.SaveSubscription(subs1, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	err = env.SaveSubscription(subs2, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag prorated orders",
			args: args{
				subs: upSubs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.MarkOrdersProrated(tt.args.subs); (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.MarkOrdersProrated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}

func TestMemberTx_UpsertMember(t *testing.T) {

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	m, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		mm paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Insert Membership",
			args: args{
				mm: test.GenMember(test.NewProfile().RandomUser(), false),
			},
			wantErr: false,
		},
		{
			name: "Update Membership",
			args: args{
				mm: test.GenMember(test.MyProfile.User(test.IDFtc), false),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.UpsertMember(tt.args.mm); (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.UpsertMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Created membership: %+v", tt.args.mm)
		})
	}

	if err := m.commit(); err != nil {
		panic(err)
	}
}
