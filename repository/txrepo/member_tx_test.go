package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	store := test.NewSubStore(test.NewProfile())

	m := store.MustGetMembership()

	test.NewRepo().
		MustSaveMembership(m)

	otx := NewMemberTx(test.DB.MustBegin(), false)

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
				id: test.NewProfile().AccountID(),
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

	otx := NewMemberTx(test.DB.MustBegin(), false)

	type args struct {
		order subs.Order
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

	order := store.MustCreateOrder()

	test.NewRepo().MustSaveOrder(order)

	otx := NewMemberTx(test.DB.MustBegin(), false)

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

	order := store.MustCreateOrder()

	test.NewRepo().
		MustSaveOrder(order)

	order = store.MustConfirmOrder(order.ID)

	otx := NewMemberTx(test.DB.MustBegin(), false)

	type args struct {
		order subs.Order
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

			if err := otx.ConfirmOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
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

	otx := NewMemberTx(test.DB.MustBegin(), false)

	type args struct {
		m subs.Membership
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

	m := store.MustGetMembership()
	test.NewRepo().
		MustSaveMembership(m)

	m.Tier = enum.TierPremium

	otx := NewMemberTx(test.DB.MustBegin(), false)

	type args struct {
		m subs.Membership
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

	test.NewRepo().MustSaveRenewalOrders(store.MustRenewN(3))

	otx := NewMemberTx(test.DB.MustBegin(), false)

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

func TestOrderTx_SaveUpgradeSchema(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	upgrade, _ := store.MustUpgrade(3)

	t.Logf("Upgrading schema id: %s", upgrade.ID)

	type args struct {
		up subs.UpgradeSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Upgrade Schema",
			args: args{
				up: upgrade,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewMemberTx(test.DB.MustBegin(), false)

			if err := tx.SaveUpgradeSchema(tt.args.up); (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("SaveUpgradeSchema() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = tx.Commit()
		})
	}
}

func TestOrderTx_ProratedOrdersUsed(t *testing.T) {
	store := test.NewSubStore(test.NewProfile())

	upgrade, _ := store.MustUpgrade(3)
	test.NewRepo().
		SaveProratedOrders(upgrade)

	t.Logf("Upgrading schema id: %s", upgrade.ID)

	otx := NewMemberTx(test.DB.MustBegin(), false)

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
