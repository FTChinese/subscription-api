package txrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestSharedTx_CreateMember(t *testing.T) {
	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create membership",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				m: reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
					Build(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.CreateMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("CreateMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestSharedTx_RetrieveMember(t *testing.T) {

	m := reader.NewMockMemberBuilder("").Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		id pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve membership",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				id: m.UserIDs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewSharedTx(tt.fields.Tx)
			got, err := tx.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
			_ = tx.Commit()
		})
	}
}

func TestSharedTx_UpdateMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).Build()
	test.NewRepo().MustSaveMembership(m)
	t.Logf("Created membership %s", m.FtcID.String)

	m.ExpireDate = chrono.DateNow()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Update membership",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				m: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.UpdateMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("UpdateMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSharedTx_DeleteMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).Build()
	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		id pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Delete membership",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				id: m.UserIDs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.DeleteMember(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestSharedTx_SaveInvoice(t *testing.T) {
	userID := uuid.New().String()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		inv invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Invoice for create",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).Build(),
			},
		},
		{
			name: "Invoice for renewal",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindRenew).
					Build(),
			},
		},
		{
			name: "Invoice for upgrade",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindUpgrade).
					Build(),
			},
		},
		{
			name: "Invoice for user-purchase addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindAddOn).
					Build(),
			},
		},
		{
			name: "Invoice for upgrade carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().WithOrderID(pkg.MustOrderID()),
			},
		},
		{
			name: "Invoice for switching to Stripe carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().
					WithStripeSubsID(faker.GenStripeSubID()),
			},
		},
		{
			name: "Invoice for switching to Apple carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().
					WithAppleTxID(faker.GenAppleSubID()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewSharedTx(tt.fields.Tx)
			if err := tx.SaveInvoice(tt.args.inv); (err != nil) != tt.wantErr {
				t.Errorf("SaveInvoice() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}
