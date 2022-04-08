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
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"testing"
)

func TestEnv_CreatePrice(t *testing.T) {
	faker.SeedGoFake()

	env := New(db.MockMySQL())

	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Recurring price",
			args: args{
				p: price.New(price.FtcCreationParams{
					Kind:    price.KindRecurring,
					Edition: price.MockEditionStdYear,
					FtcUpdateParams: price.FtcUpdateParams{
						Title:    null.StringFrom(gofakeit.Sentence(20)),
						Nickname: null.StringFrom(gofakeit.Word()),
						PeriodCount: price.ColumnYearMonthDay{
							YearMonthDay: dt.YearMonthDay{
								Years:  1,
								Months: 0,
								Days:   0,
							},
						},
						StripePriceID: faker.GenStripePriceID(),
					},

					ProductID:  ids.ProductID(),
					UnitAmount: 298,
				}, false),
			},
			wantErr: false,
		},
		{
			name: "One time price",
			args: args{
				p: price.New(price.FtcCreationParams{
					Kind: price.KindOneTime,
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleNull,
					},
					FtcUpdateParams: price.FtcUpdateParams{
						Title:    null.StringFrom(gofakeit.Sentence(2)),
						Nickname: null.StringFrom(gofakeit.Word()),
						PeriodCount: price.ColumnYearMonthDay{
							YearMonthDay: dt.YearMonthDay{
								Years:  0,
								Months: 0,
								Days:   7,
							},
						},
						StripePriceID: faker.GenStripePriceID(),
					},
					ProductID:  ids.ProductID(),
					UnitAmount: 1,
				}, false),
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

func TestEnv_UpdatePrice(t *testing.T) {
	p := test.NewStdProdBuilder().
		NewPriceBuilder(enum.CycleNull).
		Build()

	repo := test.NewRepo()

	repo.CreatePrice(p)

	env := New(db.MockMySQL())

	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update price",
			args: args{
				p: p.Update(price.FtcUpdateParams{
					Title:         null.StringFrom("Updating description"),
					Nickname:      null.StringFrom("Updating nickname"),
					StripePriceID: "changed_stripe_id",
				}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdatePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_ActivatePrice(t *testing.T) {
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

	env := New(db.MockMySQL())

	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Activate price",
			args: args{
				p: p4.Activate(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Activating %s", tt.args.p.ID)

			if err := env.ActivatePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ActivatePrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_UpdatePriceOffers(t *testing.T) {

	p := test.NewStdProdBuilder().NewYearPriceBuilder().Build()

	test.NewRepo().CreatePrice(p)

	env := New(db.MockMySQL())

	type args struct {
		pwp pw.PaywallPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update price offers",
			args: args{
				pwp: pw.PaywallPrice{
					FtcPrice: p,
					Offers: []price.Discount{
						test.NewDiscountBuilder(p.ID).BuildPromo(),
						test.NewDiscountBuilder(p.ID).BuildRetention(),
						test.NewDiscountBuilder(p.ID).BuildWinBack(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdatePriceOffers(tt.args.pwp); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePriceOffers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_ListPrices(t *testing.T) {
	prodBuilder := test.NewStdProdBuilder()

	p1 := prodBuilder.NewYearPriceBuilder().Build()
	p2 := prodBuilder.NewMonthPriceBuilder().Build()

	repo := test.NewRepo()
	repo.CreatePrice(p1)
	repo.CreatePrice(p2)

	env := New(db.MockMySQL())

	type args struct {
		prodID string
		live   bool
	}
	tests := []struct {
		name    string
		args    args
		want    []pw.PaywallPrice
		wantErr bool
	}{
		{
			name: "List prices",
			args: args{
				prodID: p1.ProductID,
				live:   false,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListProductPrices(tt.args.prodID, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProductPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListProductPrices() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ArchivePrice(t *testing.T) {

	prodBuilder := test.NewStdProdBuilder()

	p1 := prodBuilder.NewYearPriceBuilder().Build()

	repo := test.NewRepo()
	repo.CreatePrice(p1)

	env := New(db.MockMySQL())

	type args struct {
		p price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archive price",
			args: args{
				p: p1.Archive(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.ArchivePrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ArchivePrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_CreatePlan(t *testing.T) {

	env := New(db.MockMySQL())

	type args struct {
		p price.Plan
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create plan",
			args: args{
				p: price.NewPlan(test.NewStdProdBuilder().NewMonthPriceBuilder().Build()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.CreatePlan(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("CreatePlan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
