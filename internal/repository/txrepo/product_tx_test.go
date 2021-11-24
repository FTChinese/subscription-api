package txrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestProductTx_DeactivateSiblingProduct(t *testing.T) {

	prod := test.NewStdProdBuilder().Build()

	test.NewRepo().CreateProduct(prod)

	t.Logf("New product %s", faker.MustMarshalIndent(prod))

	myDBs := db.MockMySQL()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Deactivate sibling product",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := ProductTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.DeactivateSiblingProduct(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("DeactivateSiblingProduct() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			err := tx.Commit()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestProductTx_ActivateProduct(t *testing.T) {
	prod := test.NewStdProdBuilder().Build()

	test.NewRepo().CreateProduct(prod)

	t.Logf("New product %s", faker.MustMarshalIndent(prod))

	myDBs := db.MockMySQL()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Activate product",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := ProductTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.ActivateProduct(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ActivateProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
			err := tx.Commit()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestProductTx_SetProductOnPaywallLegacy(t *testing.T) {

	prod := test.NewStdProdBuilder().Build()

	test.NewRepo().CreateProduct(prod)

	t.Logf("New product %s", faker.MustMarshalIndent(prod))

	myDBs := db.MockMySQL()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Put product on paywall",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := ProductTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.SetProductOnPaywallLegacy(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SetProductOnPaywallLegacy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err := tx.Commit()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestProductTx_SetProductOnPaywall(t *testing.T) {

	prod := test.NewStdProdBuilder().Build()

	prod2 := test.NewStdProdBuilder().WithLive().Build()

	prod3 := test.NewPrmProdBuilder().Build()

	prod4 := test.NewPrmProdBuilder().WithLive().Build()

	t.Logf("New product %s", faker.MustMarshalIndent(prod))

	myDBs := db.MockMySQL()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Put product on paywall",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod,
			},
			wantErr: false,
		},
		{
			name: "Put product on paywall",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod2,
			},
			wantErr: false,
		},
		{
			name: "Put product on paywall",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod3,
			},
			wantErr: false,
		},
		{
			name: "Put product on paywall",
			fields: fields{
				Tx: myDBs.Write.MustBegin(),
			},
			args: args{
				p: prod4,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := ProductTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.SetProductOnPaywall(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SetProductOnPaywall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err := tx.Commit()
			if err != nil {
				t.Error(err)
			}
		})
	}
}
