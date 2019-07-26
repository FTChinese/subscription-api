package model

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {
	store := test.NewSubStore(test.NewProfile().AccountID(test.AccountKindFtc))
	test.NewModel().NewMemberCreated(store)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		id paywall.AccountID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty member",
			args: args{
				id: test.NewProfile().AccountID(test.AccountKindFtc),
			},
			wantErr: false,
		},
		{
			name: "Existing member",
			args: args{
				id: store.Member.User,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}
			got, err := tx.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("%+v", got)
		})
	}
}

func TestOrderTx_SaveOrder(t *testing.T) {
	p := test.NewProfile()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		s paywall.Subscription
		c util.ClientApp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "FTC account with alipay",
			args: args{
				s: test.BuildSubs(p.FtcAccountID(), enum.PayMethodAli, paywall.SubsKindCreate),
				c: test.RandomClientApp(),
			},
			wantErr: false,
		},
		{
			name: "WX account with wxpay",
			args: args{
				s: test.BuildSubs(p.WxAccountID(), enum.PayMethodWx, paywall.SubsKindRenew),
				c: test.RandomClientApp(),
			},
			wantErr: false,
		},
		{
			name: "Linked account upgrading",
			args: args{
				s: test.BuildSubs(p.LinkedAccountID(), enum.PayMethodAli, paywall.SubsKindUpgrade),
				c: test.RandomClientApp(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := otx.SaveOrder(tt.args.s, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_RetrieveOrder(t *testing.T) {
	p := test.NewProfile()
	store := test.NewSubStore(p.FtcAccountID())

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
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
			name: "Retrieve an unconfirmed order",
			args: args{
				orderID: test.NewModel().CreateNewOrder(store).ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			got, err := otx.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := otx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("%+v", got)
		})
	}
}

func TestOrderTx_ConfirmOrder(t *testing.T) {
	p := test.NewProfile()
	store := test.NewSubStore(p.FtcAccountID())

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		order paywall.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save order to be confirmed",
			args: args{
				order: test.NewModel().CreateConfirmedOrder(store),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := tx.ConfirmOrder(tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_CreateMember(t *testing.T) {
	p := test.NewProfile()
	store := test.NewSubStore(p.FtcAccountID())

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		m paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create a new member",
			args: args{
				m: test.NewModel().CreateNewMember(store),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := otx.CreateMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.CreateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Created new member %+v", tt.args.m)
		})
	}
}

func TestOrderTx_UpdateMember(t *testing.T) {
	p := test.NewProfile()
	store := test.NewSubStore(p.FtcAccountID())

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		m paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update member",
			args: args{
				m: test.NewModel().CreateUpdatedMember(store),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := otx.UpdateMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.UpdateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Update member: %+v", tt.args.m)
		})
	}
}

func TestOrderTx_FindBalanceSources(t *testing.T) {

	accountID := test.NewProfile().FtcAccountID()
	store := test.NewSubStore(accountID)

	test.NewModel().MemberRenewedN(store, 5)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		accountID paywall.AccountID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find Balance Sources",
			args: args{
				accountID: accountID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			got, err := tx.FindBalanceSources(tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.FindBalanceSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Balance sources: %+v", got)
		})
	}
}

func TestOrderTx_SaveUpgrade(t *testing.T) {
	accountID := test.NewProfile().FtcAccountID()
	store := test.NewSubStore(accountID)

	upgradeOrder, err := store.UpgradeOrder(5)
	if err != nil {
		t.Error(err)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		orderID string
		up      paywall.Upgrade
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Upgrade",
			args: args{
				orderID: upgradeOrder.ID,
				up:      store.UpgradeV1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := tx.SaveUpgrade(tt.args.orderID, tt.args.up); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SaveUpgrade() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Saved upgrade: %+v", tt.args.up)
		})
	}
}

func TestOrderTx_SaveUpgradeV2(t *testing.T) {
	accountID := test.NewProfile().FtcAccountID()
	store := test.NewSubStore(accountID)

	upgradeOrder, err := store.UpgradeOrder(5)
	if err != nil {
		t.Error(err)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		orderID string
		up      paywall.UpgradePreview
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save upgrade v2",
			args: args{
				orderID: upgradeOrder.ID,
				up:      store.UpgradeV2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := otx.SaveUpgradeV2(tt.args.orderID, tt.args.up); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SaveUpgradeV2() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_SetLastUpgradeID(t *testing.T) {
	accountID := test.NewProfile().FtcAccountID()
	store := test.NewSubStore(accountID)

	test.NewModel().UpgradeOrder(store, 5)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		up paywall.Upgrade
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set upgrade id on upgrade source order",
			args: args{
				up: store.UpgradeV1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}
			if err := tx.SetLastUpgradeID(tt.args.up); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SetLastUpgradeID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_ConfirmUpgrade(t *testing.T) {
	accountID := test.NewProfile().FtcAccountID()
	store := test.NewSubStore(accountID)

	upgradeOrder := test.NewModel().UpgradeOrder(store, 5)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	otx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	if err := otx.SaveUpgradeV2(upgradeOrder.ID, store.UpgradeV2); err != nil {
		t.Error(err)
	}

	if err := otx.SetLastUpgradeIDV2(store.UpgradeV2); err != nil {
		t.Error(err)
	}

	if err := otx.commit(); err != nil {
		t.Error(err)
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirm upgrade",
			args: args{
				id: store.UpgradeV2.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}
			if err := tx.ConfirmUpgrade(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.ConfirmUpgrade() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Confirmed upgrade id: %s", tt.args.id)
		})
	}
}
