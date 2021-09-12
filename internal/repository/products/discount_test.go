package products

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_LoadDiscount(t *testing.T) {
	env := NewEnv(db.MockMySQL(), nil)

	pb := test.NewStdProdBuilder().NewYearPriceBuilder()
	disc := pb.NewDiscountBuilder().BuildRetention()

	test.NewRepo().CreateDiscount(disc)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    price.Discount
		wantErr bool
	}{
		{
			name: "Load discount",
			args: args{
				id: disc.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.LoadDiscount(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadDiscount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("LoadDiscount() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_CreateDiscount(t *testing.T) {
	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		d price.Discount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create discount",
			args: args{
				d: test.NewStdProdBuilder().
					NewMonthPriceBuilder().
					NewDiscountBuilder().
					BuildIntro(),
			},
			wantErr: false,
		},
		{
			name: "Create discount",
			args: args{
				d: test.NewProductBuilder(enum.TierStandard).
					NewPriceBuilder(enum.CycleYear).
					NewDiscountBuilder().
					BuildPromo(),
			},
			wantErr: false,
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
	env := NewEnv(db.MockMySQL(), nil)

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
	env := NewEnv(db.MockMySQL(), nil)

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
	env := NewEnv(db.MockMySQL(), nil)

	pb := test.NewStdProdBuilder().NewYearPriceBuilder()

	pri := pb.Build()
	repo := test.NewRepo()
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildIntro())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildWinBack())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildRetention())
	repo.CreateDiscount(pb.NewDiscountBuilder().BuildPromo().Cancel())

	type args struct {
		priceID string
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListDiscounts(tt.args.priceID)
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
