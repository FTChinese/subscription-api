package repository

import (
	"testing"

	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"go.uber.org/zap/zaptest"
)

func TestStripeRepo_ListPaywallPrices(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    []price.StripePrice
		wantErr bool
	}{
		{
			name: "list active prices",
			args: args{
				live: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.ListPaywallPrices(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("StripeRepo.ListActivePrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got items %d", len(got))

			t.Logf("items %s", faker.MustMarshalIndent(got))
		})
	}
}

func TestStripeRepo_ListPaywallCoupons(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "",
			args: args{
				live: false,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.ListPaywallCoupons(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaywallCoupons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if len(got) != tt.want {
			// 	t.Errorf("RetrievePaywallCoupons() got = %d, want %d", len(got), tt.want)
			// }

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
