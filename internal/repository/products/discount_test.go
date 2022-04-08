package products

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_CreateDiscount(t *testing.T) {
	env := New(db.MockMySQL())

	type args struct {
		d price.Discount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Intro discount",
			args: args{
				d: test.NewStdProdBuilder().
					NewMonthPriceBuilder().
					NewDiscountBuilder().
					BuildIntro(),
			},
			wantErr: false,
		},
		{
			name: "Promo discount",
			args: args{
				d: test.NewProductBuilder(enum.TierStandard).
					NewPriceBuilder(enum.CycleYear).
					NewDiscountBuilder().
					BuildPromo(),
			},
			wantErr: false,
		},
		{
			name: "Custom discount period",
			args: args{
				d: test.NewDiscountBuilder(ids.DiscountID()).
					WithMode(false).
					WithPriceOff(10).
					WithPeriod(
						dt.YearMonthDay{
							Years:  0,
							Months: 3,
							Days:   0,
						}).
					BuildPromo(),
			},
		},
		{
			name: "Specified price",
			args: args{
				d: test.NewDiscountBuilder("price_WHc5ssjh6pqw").
					WithMode(false).
					WithPriceOff(20).
					WithPeriod(
						dt.YearMonthDay{
							Years:  1,
							Months: 3,
							Days:   0,
						}).
					BuildPromo(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.CreateDiscount(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("CreateDiscount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_UpdateDiscount(t *testing.T) {
	env := New(db.MockMySQL())

	pb := test.NewStdProdBuilder().NewYearPriceBuilder()
	disc := pb.NewDiscountBuilder().BuildRetention()

	test.NewRepo().CreateDiscount(disc)

	disc = disc.Cancel()

	type args struct {
		d price.Discount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update discount",
			args: args{
				d: disc,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateDiscount(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDiscount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_ListActiveDiscounts(t *testing.T) {
	env := New(db.MockMySQL())

	pb := test.NewStdProdBuilder().NewYearPriceBuilder()

	pri := pb.Build()
	repo := test.NewRepo()
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildIntro())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildWinBack())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildRetention())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildPromo().Cancel())

	type args struct {
		priceID string
		live    bool
	}
	tests := []struct {
		name    string
		args    args
		want    []price.Discount
		wantErr bool
	}{
		{
			name: "List active discounts",
			args: args{
				priceID: pri.ID,
				live:    true,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListActiveDiscounts(tt.args.priceID, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListActiveDiscounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListActiveDiscounts() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ListDiscounts(t *testing.T) {
	env := New(db.MockMySQL())

	pb := test.NewStdProdBuilder().NewYearPriceBuilder()

	pri := pb.Build()
	repo := test.NewRepo()
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildIntro())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildWinBack())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildRetention())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildPromo().Cancel())

	type args struct {
		priceID string
		live    bool
	}
	tests := []struct {
		name    string
		args    args
		want    []price.Discount
		wantErr bool
	}{
		{
			name: "List discounts",
			args: args{
				priceID: pri.ID,
				live:    pri.LiveMode,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListDiscounts(tt.args.priceID, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDiscounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListDiscounts() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ArchivePriceDiscounts(t *testing.T) {
	p := test.NewStdProdBuilder().NewYearPriceBuilder().Build()

	d1 := test.NewDiscountBuilder(p.ID).BuildPromo()
	d2 := test.NewDiscountBuilder(p.ID).BuildRetention()
	d3 := test.NewDiscountBuilder(p.ID).BuildWinBack()

	repo := test.NewRepo()
	repo.CreateDiscount(d1)
	repo.CreateDiscount(d2)
	repo.CreateDiscount(d3)

	env := New(db.MockMySQL())

	type args struct {
		p pw.PaywallPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archive discount under a price",
			args: args{
				p: pw.PaywallPrice{
					FtcPrice: price.FtcPrice{
						ID: p.ID,
					},
					Offers: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.ArchivePriceDiscounts(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ArchivePriceDiscounts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
