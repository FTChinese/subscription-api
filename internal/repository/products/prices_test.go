package products

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_CreatePrice(t *testing.T) {

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
			name: "Create price",
			args: args{
				p: test.NewStdProdBuilder().
					NewYearPriceBuilder().
					Build(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.CreatePrice(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Price created\n")
		})
	}
}

func TestEnv_UpdateFtcPrice(t *testing.T) {

	p := test.NewStdProdBuilder().NewYearPriceBuilder().Build()

	test.NewRepo().CreatePrice(p)

	env := New(db.MockMySQL())

	type args struct {
		f price.FtcPrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update price",
			args: args{
				f: p,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateFtcPrice(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("UpdateFtcPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
