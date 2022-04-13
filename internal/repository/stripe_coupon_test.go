package repository

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
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
