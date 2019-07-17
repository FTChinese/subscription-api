package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
)

// Those tests only checks whether db operations are correct.
// It does not guarantees logical correctness.
func TestOrderTx_SaveOrder(t *testing.T) {
	userID := test.MyProfile.RandomUserID()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	subs := test.SubsCreate(userID)
	if err := otx.SaveOrder(subs, test.RandomClientApp()); err != nil {
		t.Error(err)
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}

	t.Logf("Saved new order: %+s", subs.OrderID)
}

// Create a new order in db and returns it.
func createOrder(order paywall.Subscription) paywall.Subscription {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.SaveOrder(order, test.RandomClientApp()); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return order
}

func TestOrderTx_RetrieveOrder(t *testing.T) {
	// Prerequisites
	userID := test.MyProfile.RandomUserID()

	subs := createOrder(test.SubsCreate(userID))
	t.Logf("Saved new order: %+s", subs.OrderID)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// Start testing.
	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	order, err := tx.RetrieveOrder(subs.OrderID)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Retrieved order: %+v", order)
}

func TestOrderTx_ConfirmOrder(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	order, _ := createOrder(test.SubsCreate(userID)).
		Confirm(paywall.Membership{}, time.Now())
	t.Logf("Created order: %s", order.OrderID)

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
			name: "Save confirmed order for creation",
			args: args{
				order: order,
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

func confirmOrder(
	unconfirmedOrder paywall.Subscription,
	previousMember paywall.Membership,
) paywall.Subscription {

	confirmedOrder, _ := createOrder(unconfirmedOrder).
		Confirm(previousMember, time.Now())

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.ConfirmOrder(confirmedOrder); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return confirmedOrder
}

func TestOrderTx_CreateMember(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	previousMember := paywall.NewMember(userID)
	order := confirmOrder(test.SubsCreate(userID), previousMember)
	t.Logf("A confirmed order: %+v", order)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// User new order to build member
	m, err := previousMember.FromAliOrWx(order)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Member: %+v", m)

	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	if err := tx.CreateMember(m); err != nil {
		t.Error(err)
	}

	if err := tx.commit(); err != nil {
		t.Error(err)
	}
}

func newMember(userID paywall.UserID) paywall.Membership {

	previousMember := paywall.NewMember(userID)
	order := confirmOrder(test.SubsCreate(userID), previousMember)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// User new order to build member
	m, err := previousMember.FromAliOrWx(order)
	if err != nil {
		panic(err)
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.CreateMember(m); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return m
}

func TestOrderTx_UpdateMember(t *testing.T) {
	userID := test.NewProfile().UserID(test.IDFtc)

	previousMember := newMember(userID)
	t.Logf("Existing membership: %+v", previousMember)

	renewalSubs := confirmOrder(test.SubsRenew(userID), previousMember)

	renewedMember, err := previousMember.FromAliOrWx(renewalSubs)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Renewed membership: %+v", renewedMember)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	if err := tx.UpdateMember(renewedMember); err != nil {
		t.Error(err)
	}

	if err := tx.commit(); err != nil {
		t.Error(err)
	}
}

func renewMember(userID paywall.UserID, initial paywall.Membership) paywall.Membership {
	renewalSubs := confirmOrder(test.SubsRenew(userID), initial)

	renewedMember, err := initial.FromAliOrWx(renewalSubs)
	if err != nil {
		panic(err)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.UpdateMember(renewedMember); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return renewedMember
}

// Renew a membership N times.
func renewMemberN(userID paywall.UserID, count int) paywall.Membership {
	previousMember := newMember(userID)

	for i := 0; i < count; i++ {
		previousMember = renewMember(userID, previousMember)
	}

	return previousMember
}

func TestOrderTx_FindBalanceSources(t *testing.T) {

	userID := test.NewProfile().RandomUserID()
	finalMember := renewMemberN(userID, 3)

	t.Logf("Final member %+v", finalMember)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		userID paywall.UserID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find Balance Sources",
			args: args{
				userID: userID,
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

			got, err := tx.FindBalanceSources(tt.args.userID)
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

func buildUpgrade(userID paywall.UserID, count int) paywall.Upgrade {
	member := renewMemberN(userID, count)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	sources, err := tx.FindBalanceSources(userID)

	if err := tx.commit(); err != nil {
		panic(err)
	}

	up := paywall.NewUpgrade(test.YearlyPremium).SetBalance(sources).CalculatePayable()

	up.Member = member
	return up
}

func TestOrderTx_SaveUpgrade(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	upgrade := buildUpgrade(userID, 3)

	t.Logf("Upgrade %+v", upgrade)

	upgradeOrder := createOrder(test.SubsUpgrade(userID, upgrade))

	t.Logf("Created upgrade order: %+v", upgradeOrder)

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
				orderID: upgradeOrder.OrderID,
				up:      upgrade,
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

func buildUpgradeV2(userID paywall.UserID, count int) paywall.UpgradePreview {
	member := renewMemberN(userID, count)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	sources, err := tx.FindBalanceSources(userID)

	if err := tx.commit(); err != nil {
		panic(err)
	}

	up := paywall.NewUpgradePreview(sources)
	up.Plan = test.YearlyPremium.BuildUpgradePlan(up.Balance)
	up.Member = member
	return up
}

func TestOrderTx_SaveUpgradeV2(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	upgrade := buildUpgradeV2(userID, 3)

	t.Logf("Upgrade %+v", upgrade)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	orderID, _ := paywall.GenerateOrderID()

	if err := tx.SaveUpgradeV2(orderID, upgrade); err != nil {
		t.Error(err)
	}

	if err := tx.commit(); err != nil {
		t.Error(err)
	}
}

func TestOrderTx_SetUpgradeIDOnSource(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	upgrade := buildUpgrade(userID, 3)

	t.Logf("Upgrade %+v", upgrade)

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
				up: upgrade,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}
			if err := tx.SetUpgradeIDOnSource(tt.args.up); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SetUpgradeIDOnSource() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func createUpgrade(userID paywall.UserID, count int) (paywall.Upgrade, paywall.Subscription) {
	upgrade := buildUpgrade(userID, count)

	upgradeOrder := createOrder(test.SubsUpgrade(userID, upgrade))

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.SaveUpgrade(upgradeOrder.OrderID, upgrade); err != nil {
		panic(err)
	}

	if err := tx.SetUpgradeIDOnSource(upgrade); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return upgrade, upgradeOrder
}

func TestOrderTx_ConfirmUpgrade(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	upgrade, _ := createUpgrade(userID, 3)

	t.Logf("Created upgrade +%v", upgrade)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
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
				id: upgrade.ID,
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
		})
	}
}

// Test the process of upgrading a membership.
func TestOrderTx_CreateUpgradedMember(t *testing.T) {
	// Prerequisites to create an existing user, an upgrade order,
	// and an upgrade metadata.
	userID := test.NewProfile().RandomUserID()
	upgrade, order := createUpgrade(userID, 3)

	confirmedSubs, err := order.Confirm(upgrade.Member, time.Now())
	if err != nil {
		t.Error(err)
	}

	t.Logf("Confirmed upgrade order: %+v", confirmedSubs)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}

	// Set the order and upgrade as confirmed.
	if err := tx.ConfirmOrder(confirmedSubs); err != nil {
		t.Error(err)
	}

	if err := tx.ConfirmUpgrade(upgrade.ID); err != nil {
		t.Error(err)
	}

	// Build upgrade membership from previous membership
	// and upgrade order.
	upgradedMember, err := upgrade.Member.FromAliOrWx(confirmedSubs)
	if err != nil {
		t.Error(err)
	}

	// Update membership
	if err := tx.UpdateMember(upgradedMember); err != nil {
		t.Error(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}
}

func createUpgradedMember(userID paywall.UserID, count int) paywall.Membership {
	upgrade, order := createUpgrade(userID, count)

	confirmedSubs, err := order.Confirm(upgrade.Member, time.Now())
	if err != nil {
		panic(err)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	// Set the order and upgrade as confirmed.
	if err := tx.ConfirmOrder(confirmedSubs); err != nil {
		panic(err)
	}

	if err := tx.ConfirmUpgrade(upgrade.ID); err != nil {
		panic(err)
	}

	// Build upgrade membership from previous membership
	// and upgrade order.
	upgradedMember, err := upgrade.Member.FromAliOrWx(confirmedSubs)
	if err != nil {
		panic(err)
	}

	// Update membership
	if err := tx.UpdateMember(upgradedMember); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return upgradedMember
}

// Make sure if existing orders were used for upgrading,
// they will never be used again.
func TestOrderTx_RetrieveUsedBalanceSources(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	m := createUpgradedMember(userID, 3)

	t.Logf("Upgraded membership: %+v", m)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		t.Error(err)
	}
	sources, err := tx.FindBalanceSources(userID)
	if err != nil {
		t.Error(err)
	}

	if len(sources) != 0 {
		t.Errorf("Balance sources should already be used. Got %+v", sources)
	}

	t.Logf("Sources: %+v", sources)

	if err := tx.commit(); err != nil {
		t.Error(err)
	}
}
