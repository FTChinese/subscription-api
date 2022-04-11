package repository

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_SaveShoppingSession(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	sp := price.MockNewStripePrice()

	type args struct {
		s stripe.ShoppingSession
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "save shopping cart",
			args: args{
				s: stripe.ShoppingSession{
					FtcUserID: uuid.New().String(),
					RecurringPrice: stripe.PriceColumn{
						StripePrice: sp,
					},
					IntroductoryPrice: stripe.PriceColumn{},
					Membership:        reader.MembershipColumn{},
					Intent: reader.CheckoutIntent{
						Kind:  reader.IntentCreate,
						Error: nil,
					},
					RequestParams: stripe.SubsReqParamsColumn{
						SubsParams: stripe.SubsParams{
							PriceID:              sp.ID,
							IntroductoryPriceID:  null.String{},
							CouponID:             null.String{},
							DefaultPaymentMethod: null.String{},
							IdempotencyKey:       "",
						},
					},
					CreatedUTC: chrono.TimeNow(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.SaveShoppingSession(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SaveShoppingSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
