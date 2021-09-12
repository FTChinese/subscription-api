package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func TestAddOnTx_AddOnInvoices(t *testing.T) {
	userID := uuid.New().String()

	repo := test.NewRepo()
	repo.MustSaveInvoiceN([]invoice.Invoice{
		invoice.NewMockInvoiceBuilder().
			WithFtcID(userID).
			WithOrderKind(enum.OrderKindAddOn).
			Build(),
		invoice.NewMockInvoiceBuilder().
			WithFtcID(userID).
			WithOrderKind(enum.OrderKindAddOn).
			WithAddOnSource(addon.SourceCarryOver).
			Build(),
		invoice.NewMockInvoiceBuilder().
			WithFtcID(userID).
			WithOrderKind(enum.OrderKindAddOn).
			WithAddOnSource(addon.SourceCompensation).
			Build(),
	})

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		ids ids.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "List addons",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				ids: ids.NewFtcUserID(userID),
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewAddOnTx(tt.fields.Tx)
			got, err := tx.AddOnInvoices(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOnInvoices() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}
			if len(got) != tt.want {
				t.Errorf("Got slice len %d, want %d", len(got), tt.want)
				_ = tx.Rollback()
				return
			}
			_ = tx.Commit()
		})
	}
}

func TestAddOnTx_AddOnInvoiceConsumed(t *testing.T) {

	userID := uuid.New().String()

	inv1 := invoice.NewMockInvoiceBuilder().
		WithFtcID(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()
	inv2 := invoice.NewMockInvoiceBuilder().
		WithFtcID(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCarryOver).
		Build()
	inv3 := invoice.NewMockInvoiceBuilder().
		WithFtcID(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCompensation).
		Build()

	repo := test.NewRepo()
	repo.MustSaveInvoice(inv1)
	repo.MustSaveInvoice(inv2)
	repo.MustSaveInvoice(inv3)

	inv1 = inv1.SetPeriod(time.Now())
	inv2 = inv2.SetPeriod(time.Now())
	inv3 = inv3.SetPeriod(time.Now())

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
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv1,
			},
			wantErr: false,
		},
		{
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv2,
			},
			wantErr: false,
		},
		{
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewAddOnTx(tt.fields.Tx)
			if err := tx.AddOnInvoiceConsumed(tt.args.inv); (err != nil) != tt.wantErr {
				t.Errorf("AddOnInvoiceConsumed() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}
