package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestCounter_checkout(t *testing.T) {
	account := reader.MockNewFtcAccount(enum.AccountKindFtc)

	type fields struct {
		Account reader.FtcAccount
		Price   price.FtcPrice
		Method  enum.PayMethod
		WxAppID null.String
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Checkout
		wantErr bool
	}{
		{
			name: "New subscription",
			fields: fields{
				Account: account,
				Price:   faker.PriceStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				m: reader.Membership{},
			},
			want: Checkout{
				Kind: enum.OrderKindCreate,
				Cart: cart.Cart{
					Price:    faker.PriceStdYear.Price,
					Discount: faker.PriceStdYear.PromotionOffer,
				},
				Payable: price.Charge{
					Amount:   298,
					Currency: "cny",
				},
				LiveMode: true,
			},
			wantErr: false,
		},
		{
			name: "Renew subscription",
			fields: fields{
				Account: account,
				Price:   faker.PriceStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				m: reader.NewMockMemberBuilder(account.FtcID).
					Build(),
			},
			want: Checkout{
				Kind: enum.OrderKindRenew,
				Cart: cart.Cart{
					Price:    faker.PriceStdYear.Price,
					Discount: faker.PriceStdYear.PromotionOffer,
				},
				Payable: price.Charge{
					Amount:   298,
					Currency: "cny",
				},
				LiveMode: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Counter{
				Account: tt.fields.Account,
				Price:   tt.fields.Price,
				Method:  tt.fields.Method,
				WxAppID: tt.fields.WxAppID,
			}
			got, err := c.checkout(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkout() got = %v, want %v", got, tt.want)
			}
		})
	}
}
