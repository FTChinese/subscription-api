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
	subs1 := test.BuildSubs(u, enum.PayMethodWx, paywall.SubsKindRenew)
	subs2 := test.SubsUpgrade(u, []paywall.Subscription{subs1})

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	saveOrder(env, subs1)
	saveOrder(env, subs2)

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

			mtx, err := env.BeginMemberTx()
			if err != nil {
				panic(err)
			}

			got, err := mtx.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mtx.commit(); err != nil {
				panic(err)
			}

			t.Logf("MemberTx.RetireveOrder() = %+v", got)
		})
	}
}

func TestMemberTx_RetrieveMember(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	p := test.NewProfile()
	u := p.RandomUser()

	createMember(env, u)

	subs1 := test.SubsRandom(test.NewProfile().RandomUser())
	subs2 := test.SubsRandom(u)

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
		mtx, err := env.BeginMemberTx()
		if err != nil {
			panic(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := mtx.RetrieveMember(tt.args.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mtx.commit(); err != nil {
				panic(err)
			}

			t.Logf("MemberTx.RetrieveMember() = %v", got)
		})
	}
}

func TestMemberTx_InvalidUpgrade(t *testing.T) {
	u := test.MyProfile.RandomUser()
	stdSubs1 := test.SubsCreate(u)
	stdSubs2 := test.SubsRenew(u)

	subs1 := test.SubsUpgrade(u, []paywall.Subscription{stdSubs1, stdSubs2})
	subs2 := test.SubsUpgrade(u, []paywall.Subscription{stdSubs1, stdSubs2})

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	saveOrder(env, subs1)
	saveOrder(env, subs2)

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
			mtx, err := env.BeginMemberTx()
			if err != nil {
				panic(err)
			}

			err = mtx.InvalidUpgrade(tt.args.orderID, tt.args.errInvalid)

			if (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.InvalidUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mtx.commit(); err != nil {
				panic(err)
			}
		})
	}
}

func TestMemberTx_ConfirmOrder(t *testing.T) {
	p := test.NewProfile()
	u := p.RandomUser()
	subs := test.SubsRandom(u)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	saveOrder(env, subs)

	subs, _ = subs.Confirm(paywall.Membership{}, time.Now())

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

			mtx, err := env.BeginMemberTx()
			if err != nil {
				panic(err)
			}

			if err := mtx.ConfirmOrder(tt.args.subs); (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mtx.commit(); err != nil {
				panic(err)
			}
		})
	}
}

func TestMemberTx_UpsertMember(t *testing.T) {

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx, err := env.BeginMemberTx()
			if err != nil {
				panic(err)
			}

			if err := mtx.UpsertMember(tt.args.mm); (err != nil) != tt.wantErr {
				t.Errorf("MemberTx.UpsertMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mtx.commit(); err != nil {
				panic(err)
			}

			t.Logf("Created membership: %+v", tt.args.mm)
		})
	}
}
