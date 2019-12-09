package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	profile := test.NewProfile()

	test.NewRepo().SaveMember(profile.Membership(reader.AccountKindFtc))

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
				id: profile.AccountID(reader.AccountKindFtc),
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

	order := store.MustCreateOrder()
	test.NewRepo().SaveOrder(order)

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

	order := store.MustCreateOrder()
	test.NewRepo().SaveOrder(order)

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

	order := store.MustCreateOrder()
	store.MustConfirmOrder(order.ID)

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
				m: store.Member,
			},
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

	order := store.MustCreateOrder()
	store.MustConfirmOrder(order.ID)
	test.NewRepo().SaveMember(store.Member)

	store.Member.Tier = enum.TierPremium

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
				m: store.Member,
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
