package shared

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_LoadDiscount(t *testing.T) {
	env := NewPaywallCommon(db.MockMySQL(), nil)

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
