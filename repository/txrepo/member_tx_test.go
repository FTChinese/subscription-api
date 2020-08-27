package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	p := test.NewPersona()

	repo := test.NewRepo()
	repo.MustSaveMembership(p.Membership())

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

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
				id: test.NewPersona().AccountID(),
			},
		},
		{
			name: "Retrieve existing member",
			args: args{
				id: p.AccountID(),
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

			t.Logf("Got: %+v", got.CompoundID)
		})
	}

	_ = otx.Commit()
}

func TestMemberTx_RetrieveAppleMember(t *testing.T) {
	p := test.NewPersona()
	m := p.SetPayMethod(enum.PayMethodApple).Membership()

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	tx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		transactionID string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Retrieve an IAP member",
			args: args{
				transactionID: m.AppleSubsID.String,
			},
			want:    p.FtcID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tx.RetrieveAppleMember(tt.args.transactionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, got.CompoundID, tt.want)
		})
	}

	_ = tx.Commit()
}

func TestOrderTx_SaveOrder(t *testing.T) {

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		order subs.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Ali Order",
			args: args{
				order: test.NewPersona().CreateOrder(),
			},
			wantErr: false,
		},
		{
			name: "Save wx order",
			args: args{
				order: test.NewPersona().SetPayMethod(enum.PayMethodWx).CreateOrder(),
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
	p := test.NewPersona()
	orderAli := p.CreateOrder()
	orderWx := p.SetPayMethod(enum.PayMethodWx).CreateOrder()

	test.NewRepo().MustSaveOrder(orderAli)
	test.NewRepo().MustSaveOrder(orderWx)

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

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
				orderID: p.CreateOrder().ID,
			},
			wantErr: true,
		},
		{
			name: "Retrieve ali order",
			args: args{
				orderID: orderAli.ID,
			},
			wantErr: false,
		},
		{
			name: "Retrieve wx order",
			args: args{
				orderID: orderWx.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.RetrieveOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Order ID: %s", got.ID)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_ConfirmedOrder(t *testing.T) {
	p := test.NewPersona()
	order := p.CreateOrder()

	test.NewRepo().
		MustSaveOrder(order)

	confirmed := p.ConfirmOrder(order)

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

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
				order: confirmed.Order,
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

			t.Logf("Confirmed order ID: %s", tt.args.order.ID)
			_ = otx.Commit()
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_CreateMember(t *testing.T) {
	p := test.NewPersona()

	member := p.Membership()

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		m reader.Membership
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

			t.Logf("Saved membership: %s", tt.args.m.CompoundID)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_UpdateMember(t *testing.T) {
	p := test.NewPersona()
	m := p.Membership()

	test.NewRepo().MustSaveMembership(m)

	m.Tier = enum.TierPremium

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		m reader.Membership
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

			t.Logf("Updated member id: %s", tt.args.m.CompoundID)
		})
	}

	_ = otx.Commit()
}

func TestMemberTx_DeleteMember(t *testing.T) {
	p := test.NewPersona()
	m := p.Membership()

	test.NewRepo().MustSaveMembership(m)

	tx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Delete a member",
			args: args{
				id: p.AccountID(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tx.DeleteMember(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderTx_FindBalanceSources(t *testing.T) {
	p := test.NewPersona()

	repo := test.NewRepo()

	for i := 0; i < 3; i++ {
		order := p.CreateOrder()
		c := p.ConfirmOrder(order)
		repo.MustSaveOrder(c.Order)
	}

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

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
				accountID: p.AccountID(),
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

func TestMemberTx_SaveProratedOrders(t *testing.T) {

	tx, _ := test.DB.Beginx()

	type fields struct {
		Tx     *sqlx.Tx
		dbName config.SubsDB
	}
	type args struct {
		po []subs.ProratedOrder
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save proration",
			fields: fields{
				Tx:     tx,
				dbName: config.SubsDBProd,
			},
			args: args{
				po: test.GenProratedOrders(subs.MustGenerateOrderID()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx:     tt.fields.Tx,
				dbName: tt.fields.dbName,
			}
			if err := tx.SaveProratedOrders(tt.args.po); (err != nil) != tt.wantErr {
				t.Errorf("SaveProratedOrders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderTx_ProratedOrdersUsed(t *testing.T) {
	upOrderID := subs.MustGenerateOrderID()

	pos := test.GenProratedOrders(upOrderID)

	test.NewRepo().
		MustSaveProratedOrders(pos)

	otx := NewMemberTx(test.DB.MustBegin(), test.CFG)

	type args struct {
		upOrderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag prorated orders as used",
			args: args{
				upOrderID: upOrderID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.ProratedOrdersUsed(tt.args.upOrderID); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("ProratedOrdersUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	_ = otx.Commit()
}
