package products

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_CreatePrice(t *testing.T) {
	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create ftc price live mode",
			args: args{
				p: price.FtcPrice{
					Price: test.
						NewStdProdBuilder().
						NewYearPriceBuilder().
						Build(),
					Offers: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Create ftc price test mode",
			args: args{
				p: price.FtcPrice{
					Price: test.
						NewStdProdBuilder().
						NewYearPriceBuilder().
						WithTest().
						Build(),
					Offers: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.CreatePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("CreatePrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_ActivatePrice(t *testing.T) {

	prodBuilder := test.NewProductBuilder(enum.TierStandard)

	p1 := prodBuilder.NewYearPriceBuilder().
		Build()
	p2 := prodBuilder.NewYearPriceBuilder().
		Build()
	p3 := prodBuilder.NewYearPriceBuilder().
		WithTest().
		WithActive().
		Build()

	test.NewRepo().CreatePrice(p1)
	test.NewRepo().CreatePrice(p2)
	test.NewRepo().CreatePrice(p3)

	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
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
			name: "Activate price",
			args: args{
				id: p1.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.ActivatePrice(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ActivatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_RetrieveFtcPrice(t *testing.T) {
	repo := test.NewRepo()
	p := test.NewStdProdBuilder().NewYearPriceBuilder().Build()
	repo.CreatePrice(p)

	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    price.FtcPrice
		wantErr bool
	}{
		{
			name: "Retrieve ftc price",
			args: args{
				id: p.ID,
			},
			want:    price.FtcPrice{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveFtcPrice(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveFtcPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveFtcPrice() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_UpdateFtcPriceOffers(t *testing.T) {

	priceBuilder := test.NewStdProdBuilder().
		NewYearPriceBuilder()

	stdYearPrice := priceBuilder.Build()

	offers := []price.Discount{
		priceBuilder.NewDiscountBuilder().BuildPromo(),
		priceBuilder.NewDiscountBuilder().BuildRetention(),
		priceBuilder.NewDiscountBuilder().BuildWinBack(),
	}

	test.NewRepo().CreatePrice(stdYearPrice)

	ftcPrice := price.FtcPrice{
		Price:  stdYearPrice,
		Offers: offers,
	}

	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		f price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set offers to price",
			args: args{
				f: ftcPrice,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateFtcPriceOffers(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("UpdateFtcPriceOffers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RefreshFtcPriceOffers(t *testing.T) {
	priceBuilder := test.NewStdProdBuilder().
		NewYearPriceBuilder()

	stdYearPrice := priceBuilder.Build()

	repo := test.NewRepo()
	repo.CreatePrice(stdYearPrice)
	repo.CreateDiscount(priceBuilder.NewDiscountBuilder().BuildPromo())
	repo.CreateDiscount(priceBuilder.NewDiscountBuilder().BuildRetention())
	repo.CreateDiscount(priceBuilder.NewDiscountBuilder().BuildWinBack())

	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		f price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		want    price.FtcPrice
		wantErr bool
	}{
		{
			name: "Refresh ftc price offers",
			args: args{
				f: price.FtcPrice{
					Price:  stdYearPrice,
					Offers: nil,
				},
			},
			want:    price.FtcPrice{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RefreshFtcPriceOffers(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshFtcPriceOffers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RefreshFtcPriceOffers() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ListPrices(t *testing.T) {
	repo := test.NewRepo()
	pb := test.NewStdProdBuilder()
	prod := pb.Build()

	repo.CreatePrice(pb.NewYearPriceBuilder().Build())
	repo.CreatePrice(pb.NewMonthPriceBuilder().Build())
	repo.CreatePrice(pb.NewYearPriceBuilder().WithActive().Build())

	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		prodID string
	}
	tests := []struct {
		name    string
		args    args
		want    []price.FtcPrice
		wantErr bool
	}{
		{
			name: "List prices",
			args: args{
				prodID: prod.ID,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListPrices(tt.args.prodID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListPrices() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
