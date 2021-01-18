package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	p := test.NewPersona()

	repo := test.NewRepo()
	repo.MustSaveMembership(p.Membership())

	otx := NewMemberTx(test.DB.MustBegin())

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

	tx := NewMemberTx(test.DB.MustBegin())

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

func TestMemberTx_RetrieveStripeMember(t *testing.T) {
	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		subID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve stripe member",
			args: args{
				subID: "sub_IY75arTimVigIr",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tx.RetrieveStripeMember(tt.args.subID)
			if (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("RetrieveStripeMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = tx.Commit()

			if got.IsZero() {
				t.Error("Stripe membership not retrieved!")
				return
			}

			t.Logf("%v", got)
		})
	}
}

func TestOrderTx_SaveOrder(t *testing.T) {

	otx := NewMemberTx(test.DB.MustBegin())

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

func TestMemberTx_LockOrder(t *testing.T) {
	p := test.NewPersona()
	orderAli := p.CreateOrder()

	test.NewRepo().MustSaveOrder(orderAli)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Lock order",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				orderID: orderAli.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			got, err := tx.LockOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LockOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Locked order: %v", got)

			if err := tx.Commit(); err != nil {
				t.Error(err)
			}

		})
	}
}

func TestOrderTx_ConfirmedOrder(t *testing.T) {
	p := test.NewPersona()
	order := p.CreateOrder()

	test.NewRepo().
		MustSaveOrder(order)

	confirmed := p.ConfirmOrder(order)

	otx := NewMemberTx(test.DB.MustBegin())

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

	otx := NewMemberTx(test.DB.MustBegin())

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

	otx := NewMemberTx(test.DB.MustBegin())

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

func TestMemberTx_RetrieveAppleSubs(t *testing.T) {

	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		origTxID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve apple subscription",
			args: args{
				origTxID: "320000437711395",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tx.RetrieveAppleSubs(tt.args.origTxID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})

		_ = tx.Commit()
	}
}

func TestMemberTx_LinkAppleSubs(t *testing.T) {
	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set ftc id to iap subscription",
			args: args{
				link: apple.LinkInput{
					FtcID:        uuid.New().String(),
					OriginalTxID: "320000437711395",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tx.LinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("LinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

		_ = tx.Commit()
	}
}

func TestMemberTx_UnlinkAppleSubs(t *testing.T) {

	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Unset ftc id from iap subscription",
			args: args{
				link: apple.LinkInput{
					FtcID:        uuid.New().String(),
					OriginalTxID: "320000437711395",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tx.UnlinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("UnlinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = tx.Commit()
		})
	}
}
