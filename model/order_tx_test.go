package model

import (
	"testing"
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
)

func createUpgrade(userID paywall.UserID) paywall.Upgrade {
	upgrade := test.GenUpgrade(userID)
	orderID, _ := paywall.GenerateOrderID()
	upgrade.Member = test.GenMember(userID, false)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.SaveUpgrade(orderID, upgrade); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return upgrade
}

func createMember(userID paywall.UserID) paywall.Membership {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	m := test.GenMember(userID, false)

	err = tx.CreateMember(m, null.String{})
	if err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return m
}

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

func createConfirmedOrder(
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
	order := createConfirmedOrder(test.SubsCreate(userID), previousMember)
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

	if err := tx.CreateMember(m, null.String{}); err != nil {
		t.Error(err)
	}

	if err := tx.commit(); err != nil {
		t.Error(err)
	}
}

func createNewMember(userID paywall.UserID) paywall.Membership {

	previousMember := paywall.NewMember(userID)
	order := createConfirmedOrder(test.SubsCreate(userID), previousMember)

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

	if err := tx.CreateMember(m, null.String{}); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return m
}

func TestOrderTx_UpdateMember(t *testing.T) {
	userID := test.NewProfile().UserID(test.IDFtc)

	previousMember := createNewMember(userID)
	t.Logf("Existing membership: %+v", previousMember)

	renewalSubs := createConfirmedOrder(test.SubsRenew(userID), previousMember)

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

func createRenewedMember(userID paywall.UserID) paywall.Membership {
	previousMember := createNewMember(userID)

	renewalSubs := createConfirmedOrder(test.SubsRenew(userID), previousMember)

	renewedMember, err := previousMember.FromAliOrWx(renewalSubs)
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

func TestOrderTx_FindBalanceSources(t *testing.T) {

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	userID := test.NewProfile().RandomUserID()

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

func TestOrderTx_SaveUpgrade(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	orderID, _ := paywall.GenerateOrderID()

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
				orderID: orderID,
				up:      test.GenUpgrade(userID),
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
		})
	}
}

func TestOrderTx_SetUpgradeIDOnSource(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	order1 := test.SubsCreate(userID)
	order2 := test.SubsRenew(userID)

	for _, o := range []paywall.Subscription{order1, order2} {
		tx, err := env.BeginOrderTx()
		if err != nil {
			t.Error(err)
		}

		if err := tx.SaveOrder(o, test.RandomClientApp()); err != nil {
			_ = tx.rollback()
			t.Error()
		}

		if err := tx.commit(); err != nil {
			t.Error(err)
		}
	}

	up := paywall.NewUpgrade(test.YearlyPremium)
	up.Source = []string{order1.OrderID, order2.OrderID}

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
				up: up,
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

func TestOrderTx_ConfirmUpgrade(t *testing.T) {
	userID := test.NewProfile().RandomUserID()

	upgrade := createUpgrade(userID)

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

func TestOrderTx_LinkUser(t *testing.T) {
	p := test.NewProfile()

	model := test.NewModel(test.DB)
	if err := model.CreateFtcUser(p); err != nil {
		t.Error(err)
	}
	t.Logf("Created ftc user: %s", p.FtcID)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	if err := env.SaveWxUser(p.WxInfo()); err != nil {
		t.Error(err)
	}
	t.Logf("Wx user: %s", p.UnionID)

	type args struct {
		m paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Link FTC",
			args: args{
				m: test.GenMember(p.UserID(test.IDFtc), false),
			},
			wantErr: false,
		},
		{
			name: "Link Wx",
			args: args{
				m: test.GenMember(p.UserID(test.IDWx), false),
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

			if err := tx.LinkUser(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.LinkUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tx.commit(); err != nil {
				t.Error(err)
			}
		})
	}
}
