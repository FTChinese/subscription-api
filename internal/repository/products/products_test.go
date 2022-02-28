package products

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"testing"
)

func TestEnv_ListProducts(t *testing.T) {

	env := New(db.MockMySQL())

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    []pw.Product
		wantErr bool
	}{
		{
			name: "List products",
			args: args{
				live: false,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListProducts(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListProducts() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_CreateProduct(t *testing.T) {

	env := New(db.MockMySQL())

	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create product",
			args: args{
				p: test.NewStdProdBuilder().Build(),
			},
			wantErr: false,
		},
		{
			name: "Create product",
			args: args{
				p: test.NewPrmProdBuilder().Build(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.CreateProduct(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("CreateProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveProduct(t *testing.T) {
	env := New(db.MockMySQL())

	type args struct {
		id   string
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    pw.Product
		wantErr bool
	}{
		{
			name: "Retrieve product",
			args: args{
				id:   "prod_GIvJj8oo3Dmf",
				live: true,
			},
			want:    pw.Product{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveProduct(tt.args.id, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveProduct() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_UpdateProduct(t *testing.T) {
	env := New(db.MockMySQL())

	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update product",
			args: args{
				p: pw.Product{
					ID: "prod_RYRkuSDr64fv",
					ProductParams: pw.ProductParams{
						CreatedBy:   "Koepp1854",
						Description: null.StringFrom("Cast communication inflation test state advance write talk express scheme.\nPut change decade philosophy move foundation shall say instance bridge.\nEnhance introduce hear 're hell be institution require own protection.\nFunction risk background finance alter deserve leadership would structure take."),
						Heading:     "高级会员update",
						SmallPrint:  null.String{},
						Tier:        enum.TierPremium,
					},
					Active:     false,
					LiveMode:   false,
					CreatedUTC: chrono.TimeNow(),
					UpdatedUTC: chrono.TimeNow(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateProduct(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpdateProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SetProductOnPaywall(t *testing.T) {

	prod := test.NewStdProdBuilder().WithLive().Build()
	test.NewRepo().CreateProduct(prod)

	env := New(db.MockMySQL())

	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Put product on paywall",
			args: args{
				p: prod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("%s", faker.MustMarshalIndent(tt.args.p))

			if err := env.SetProductOnPaywall(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SetProductOnPaywall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SetProductIntro(t *testing.T) {
	env := New(db.MockMySQL())

	pb := test.NewStdProdBuilder()

	prod := pb.Build()

	test.NewRepo().CreateProduct(prod)

	type args struct {
		p pw.Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set intro",
			args: args{
				p: prod.WithIntroPrice(
					pb.
						NewYearPriceBuilder().
						WithOneTime().
						Build()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SetProductIntro(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SetProductIntro() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
