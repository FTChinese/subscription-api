package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	store := test.NewSubStore(test.NewProfile())

	test.NewRepo(store).
		MustSaveMembership()

	env := SubEnv{db: test.DB}

	otx, _ := env.BeginOrderTx()

	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve Empty member",
			args: args{
				id: test.NewProfile().AccountID(reader.AccountKindFtc),
			},
		},
		{
			name: "Retrieve existing member",
			args: args{
				id: store.GetMemberID(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Got: %+v", got)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_SaveOrder(t *testing.T) {

	store := test.NewSubStore(test.NewProfile())

	env := SubEnv{db: test.DB}

	otx, _ := env.BeginOrderTx()

	type args struct {
		order subscription.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Order",
			args: args{
				order: store.MustCreateOrder(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.SaveOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Saved order %s", tt.args.order.ID)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_RetrieveOrder(t *testing.T) {

	store := test.NewSubStore(test.NewProfile())

	order := test.NewRepo(store).MustCreateOrder()

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve empty order",
			args: args{
				orderID: test.MustGenOrderID(),
			},
			wantErr: true,
		},
		{
			name: "Retrieve order",
			args: args{
				orderID: order.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_UpdateConfirmedOrder(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	order := test.NewRepo(store).
		MustCreateOrder()

	order = store.MustConfirmOrder(order.ID)

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		order subscription.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update Confirmed Order",
			args: args{
				order: order,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.UpdateConfirmedOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("UpdateConfirmedOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Confirmed: %+v", tt.args.order)
			_ = otx.Commit()
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_CreateMember(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	member := store.MustGetMembership()

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		m subscription.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save membership",
			args: args{
				m: member,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.CreateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("CreateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Saved membership: %s", tt.args.m.ID.String)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_UpdateMember(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	m := test.NewRepo(store).
		MustSaveMembership()

	m.Tier = enum.TierPremium

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		m subscription.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update membership",
			args: args{
				m: m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.UpdateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("UpdateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Updated member id: %s", tt.args.m.ID.String)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_FindBalanceSources(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	test.NewRepo(store).MustRenewN(3)

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		accountID reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find balance sources",
			args: args{
				accountID: store.GetMemberID(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.FindBalanceSources(tt.args.accountID)
			if (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("FindBalanceSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Balance sources: %+v", got)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_SaveProratedOrders(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	upgrade, order := store.MustUpgrade(3)
	t.Logf("Upgrading order: %+v", order)

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		p []subscription.ProratedOrderSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Prorated Orders",
			args: args{
				p: upgrade.Sources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.SaveProratedOrders(tt.args.p); (err != nil) != tt.wantErr {
				_ = otx.Rollback()

				t.Errorf("SaveProratedOrders() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = otx.Commit()
		})
	}
}

func TestOrderTx_SaveUpgradeBalance(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	upgrade, order := store.MustUpgrade(3)
	t.Logf("Upgrading order: %+v", order)

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		up subscription.UpgradeBalanceSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save upgrade balance",
			args: args{
				up: upgrade.UpgradeBalanceSchema,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.SaveUpgradeBalance(tt.args.up); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("SaveUpgradeBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		_ = otx.Commit()
	}
}

func TestOrderTx_ProratedOrdersUsed(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	upgrade := test.NewRepo(store).SaveProratedOrders(3)

	t.Logf("Upgrading schema id: %s", upgrade.ID)

	env := SubEnv{db: test.DB}
	otx, _ := env.BeginOrderTx()

	type args struct {
		upgradeID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag prorated orders as used",
			args: args{
				upgradeID: upgrade.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.ProratedOrdersUsed(tt.args.upgradeID); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("ProratedOrdersUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	_ = otx.Commit()
}
