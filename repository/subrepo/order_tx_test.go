package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {
	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	store.MustConfirm(store.MustCreate(test.YearlyStandard).ID)

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
			name: "Empty member",
			args: args{
				id: test.NewProfile().AccountID(reader.AccountKindFtc),
			},
			wantErr: false,
		},
		{
			name: "Existing member",
			args: args{
				id: store.Member.MemberID,
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

			if err := tx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("%+v", got)
		})
	}
}

func TestOrderTx_SaveOrder(t *testing.T) {
	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		order subscription.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "FTC account new order",
			args: args{
				order: store.MustCreate(test.YearlyStandard),
			},
			wantErr: false,
		},
		{
			name: "FTC account renew order",
			args: args{
				order: store.MustRenewal(test.YearlyStandard),
			},
			wantErr: false,
		},
		{
			name: "FTC account upgrading",
			args: args{
				order: store.MustUpgrading(),
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

			if err := otx.SaveOrder(tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_RetrieveOrder(t *testing.T) {

	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	order1 := store.MustCreate(test.YearlyStandard)
	test.NewRepo().SaveOrder(order1)

	order2 := store.MustConfirm(store.MustCreate(test.YearlyStandard).ID)
	test.NewRepo().SaveOrder(order2)

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
				orderID: order1.ID,
			},
			wantErr: false,
		},
		{
			name: "Retrieve a confirmed order",
			args: args{
				orderID: order2.ID,
			},
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

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("%+v", got)
		})
	}
}

func TestOrderTx_ConfirmOrder(t *testing.T) {

	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	order := store.MustCreate(test.YearlyStandard)
	test.NewRepo().SaveOrder(order)

	order = store.MustConfirm(order.ID)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		order subscription.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirm an order",
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

			if err := tx.Commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_CreateMember(t *testing.T) {

	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	store.MustConfirm(
		store.MustCreate(test.YearlyStandard).ID,
	)

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
			name: "Create a new member",
			args: args{
				m: store.Member,
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

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Created new member %+v", tt.args.m)
		})
	}
}

func TestOrderTx_UpdateMember(t *testing.T) {

	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)

	// Create a new order and confirm it.
	store.MustConfirm(store.MustCreate(test.YearlyStandard).ID)
	test.NewRepo().SaveMember(store.Member)

	// Create another order and confirm it.
	store.MustConfirm(store.MustCreate(test.YearlyStandard).ID)

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
			name: "Update member",
			args: args{
				m: store.Member,
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

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Update member: %+v", tt.args.m)
		})
	}
}

func TestOrderTx_FindBalanceSources(t *testing.T) {

	profile := test.NewProfile()

	store := test.NewSubStore(
		profile,
		reader.AccountKindFtc,
	)

	orders := store.MustRenewN(test.YearlyStandard, 5)

	repo := test.NewRepo()
	for _, v := range orders {
		repo.SaveOrder(v)
	}

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		accountID reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find Balance Sources",
			args: args{
				accountID: store.AccountID,
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

			if err := tx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Balance sources: %+v", got)
		})
	}
}

func TestOrderTx_SaveUpgradePlan(t *testing.T) {

	store := test.NewSubStore(
		test.NewProfile(),
		reader.AccountKindFtc,
	)
	store.MustRenewN(test.YearlyStandard, 3)
	store.MustCreate(test.YearlyPremium)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		up plan.UpgradePlan
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Upgrade Plan",
			args: args{
				up: store.UpgradePlan,
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
			if err := otx.SaveUpgradePlan(tt.args.up); (err != nil) != tt.wantErr {
				t.Errorf("SaveUpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}

			t.Logf("Save upgrade plan: %+v", tt.args.up)
		})
	}
}

func TestOrderTx_SaveProration(t *testing.T) {
	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)

	store.MustRenewN(test.YearlyStandard, 3)
	store.MustCreate(test.YearlyPremium)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		p []plan.ProrationSource
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Proration",
			args: args{
				p: store.UpgradePlan.Data,
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

			if err := otx.SaveProration(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SaveProration() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOrderTx_ConfirmUpgrade(t *testing.T) {
	store := test.NewSubStore(test.NewProfile(), reader.AccountKindFtc)

	store.MustRenewN(test.YearlyStandard, 3)
	store.MustCreate(test.YearlyPremium)

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
		wantErr bool
	}{
		{
			name:    "Confirm Upgrade",
			args:    args{upgradeID: store.UpgradePlan.ID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otx, err := env.BeginOrderTx()
			if err != nil {
				t.Error(err)
			}

			if err := otx.SaveProration(store.UpgradePlan.Data); err != nil {
				t.Error(err)
			}

			if err := otx.ConfirmUpgrade(tt.args.upgradeID); (err != nil) != tt.wantErr {
				t.Errorf("ConfirmUpgrade() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := otx.Commit(); err != nil {
				t.Error(err)
			}
		})
	}
}
