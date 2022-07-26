package repository

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_RetrievePaywallPrices(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	p1 := price.MockRandomStripePrice()
	p2 := price.MockRandomStripePrice()
	p3 := price.MockRandomStripePrice()

	testRepo := test.NewRepo()
	testRepo.SaveStripePrice(p1)
	testRepo.SaveStripePrice(p2)
	testRepo.SaveStripePrice(p3)

	type args struct {
		ids  []string
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
				ids:  []string{p1.ID, p2.ID, p3.ID},
				live: false,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrievePaywallPrices(tt.args.ids, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaywallPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("RetrievePaywallPrices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripeRepo_RetrievePaywallCoupons(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	p1 := price.MockRandomStripePrice()
	p2 := price.MockRandomStripePrice()

	testRepo := test.NewRepo()
	testRepo.SaveStripeCoupons(p1.MockRandomCouponN(2))
	testRepo.SaveStripeCoupons(p2.MockRandomCouponN(1))

	type args struct {
		priceIDs []string
		live     bool
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
				priceIDs: []string{
					p1.ID,
					p2.ID,
				},
				live: false,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrievePaywallCoupons(tt.args.priceIDs, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaywallCoupons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("RetrievePaywallCoupons() got = %d, want %d", len(got), tt.want)
			}
		})
	}
}
