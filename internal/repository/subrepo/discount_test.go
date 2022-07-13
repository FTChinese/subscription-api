package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestEnv_InsertDiscountRedeemed(t *testing.T) {
	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		r ftcpay.DiscountRedeemed
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Insert discount redeemed",
			args: args{
				r: ftcpay.MockDiscountRedeemed(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.InsertDiscountRedeemed(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("InsertDiscountRedeemed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveDiscountRedeemed(t *testing.T) {
	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	redeemed := ftcpay.MockDiscountRedeemed()
	_ = env.InsertDiscountRedeemed(redeemed)
	type args struct {
		userIDs    ids.UserIDs
		discountID string
	}
	tests := []struct {
		name    string
		args    args
		want    ftcpay.DiscountRedeemed
		wantErr bool
	}{
		{
			name: "Discount redeemed",
			args: args{
				userIDs: ids.UserIDs{
					FtcID: null.StringFrom(redeemed.CompoundID),
				},
				discountID: redeemed.DiscountID,
			},
			want:    redeemed,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveDiscountRedeemed(tt.args.userIDs, tt.args.discountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveDiscountRedeemed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveDiscountRedeemed() got = %v, want %v", got, tt.want)
			}
		})
	}
}
