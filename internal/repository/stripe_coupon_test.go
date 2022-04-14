package repository

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestStripeRepo_UpsertCoupon(t *testing.T) {

	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		c price.StripeCoupon
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert a coupon",
			args: args{
				c: price.MockRandomStripeCoupon(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpsertCoupon(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("UpsertCoupon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrieveCoupon(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	c := price.MockRandomStripeCoupon()

	test.NewRepo().SaveStripeCoupon(c)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    price.StripeCoupon
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: c.ID,
			},
			want:    c,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.RetrieveCoupon(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveCoupon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveCoupon() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripeRepo_RetrieveActiveCouponsOfPrice(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	coupons := price.MockRandomCouponList(3)

	test.NewRepo().SaveStripeCoupons(coupons)

	type args struct {
		priceID string
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
				priceID: coupons[0].PriceID.String,
			},
			want:    len(coupons),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.RetrieveActiveCouponsOfPrice(tt.args.priceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveActiveCouponsOfPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("RetrieveActiveCouponsOfPrice() got = %d, want %d", len(got), tt.want)
			}
		})
	}
}
