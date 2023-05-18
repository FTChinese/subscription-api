package txrepo

import (
	"testing"

	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
)

func TestPriceTx_DeactivateSiblingPrice(t *testing.T) {
	prodBuilder := test.
		NewStdProdBuilder()

	p1 := prodBuilder.
		NewYearPriceBuilder().
		Build()
	p2 := prodBuilder.
		NewMonthPriceBuilder().
		WithActive().
		Build()
	p3 := prodBuilder.
		NewYearPriceBuilder().
		WithActive().
		Build()
	p4 := prodBuilder.
		NewYearPriceBuilder().
		Build()

	repo := test.NewRepo()

	repo.CreatePrice(p1)
	repo.CreatePrice(p2)
	repo.CreatePrice(p3)
	repo.CreatePrice(p4)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Deactivate sibling price",
			fields: fields{
				Tx: db.MockMySQL().Write.MustBegin(),
			},
			args: args{
				p: p4.Activate(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := PriceTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.DeactivateFtcSiblingPrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("DeactivateSiblingPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriceTx_ActivatePrice(t *testing.T) {
	p := test.NewStdProdBuilder().NewYearPriceBuilder().Build()

	test.NewRepo().CreatePrice(p)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Activate price",
			fields: fields{
				Tx: db.MockMySQL().Write.MustBegin(),
			},
			args: args{
				p: p.Activate(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := PriceTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.ActivateFtcPrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ActivatePrice() error = %v, wantErr %v", err, tt.wantErr)

				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestPriceTx_UpsertActivePrice(t *testing.T) {
	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p price.ActivePrice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Insert active price",
			fields: fields{
				Tx: db.MockMySQL().Write.MustBegin(),
			},
			args: args{
				p: price.MockFtcStdIntroPrice.ActiveEntry(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := PriceTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.UpsertActivePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("PriceTx.UpsertActivePrice() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestPriceTx_RemoveActivePrice(t *testing.T) {
	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p price.ActivePrice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Insert active price",
			fields: fields{
				Tx: db.MockMySQL().Write.MustBegin(),
			},
			args: args{
				p: price.MockFtcStdIntroPrice.ActiveEntry(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := PriceTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.RemoveActivePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("PriceTx.RemoveActivePrice() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}
