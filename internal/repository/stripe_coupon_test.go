package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
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
		id   string
		live bool
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
				id:   c.ID,
				live: false,
			},
			want:    c,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.RetrieveCoupon(tt.args.id, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveCoupon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveCoupon() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}

func TestStripeRepo_ListPriceCoupons(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	coupons := price.MockRandomCouponList(3)

	test.NewRepo().SaveStripeCoupons(coupons)

	type args struct {
		priceID    string
		activeOnly bool
		live       bool
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "All coupons",
			args: args{
				priceID:    coupons[0].PriceID.String,
				activeOnly: false,
				live:       false,
			},
			want:    len(coupons),
			wantErr: false,
		},
		{
			name: "Active coupons",
			args: args{
				priceID:    coupons[0].PriceID.String,
				activeOnly: true,
			},
			want:    len(coupons),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListPriceCoupons(tt.args.priceID, tt.args.activeOnly, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPriceCoupons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("ListPriceCoupons() got = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestStripeRepo_InsertCouponRedeemed(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		r stripe.CouponRedeemed
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				r: stripe.MockCouponRedeemed(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.InsertCouponRedeemed(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("InsertCouponRedeemed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_LatestCouponApplied(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))
	cr := stripe.MockCouponRedeemed()

	_ = repo.InsertCouponRedeemed(cr)

	type args struct {
		invoiceID string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.CouponRedeemed
		wantErr bool
	}{
		{
			name: "",
			args: args{
				invoiceID: cr.InvoiceID,
			},
			want:    cr,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.LatestCouponApplied(tt.args.invoiceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("InvoiceHasCouponApplied() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InvoiceHasCouponApplied() got = %v, want %v", got, tt.want)
			}
		})
	}
}
