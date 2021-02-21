package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewCheckedItem(t *testing.T) {
	type args struct {
		ep price.FtcPrice
	}
	tests := []struct {
		name string
		args args
		want cart.Cart
	}{
		{
			name: "Checkout item",
			args: args{
				ep: faker.PriceStdYear,
			},
			want: cart.Cart{
				Price:    faker.PriceStdYear.Original,
				Discount: faker.PriceStdYear.PromotionOffer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cart.NewFtcCart(tt.args.ep)
			assert.Equal(t, got, tt.want)
			t.Logf("Checkout item %+v", got)
		})
	}
}

func TestCheckedItem_Payable(t *testing.T) {
	type fields struct {
		Price    price.Price
		Discount price.Discount
	}
	tests := []struct {
		name   string
		fields fields
		want   price.Charge
	}{
		{
			name: "Payable",
			fields: fields{
				Price:    faker.PriceStdYear.Original,
				Discount: faker.PriceStdYear.PromotionOffer,
			},
			want: price.Charge{
				Amount:   258,
				Currency: "cny",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := cart.Cart{
				Price:    tt.fields.Price,
				Discount: tt.fields.Discount,
			}
			if got := i.Payable(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaymentConfig_checkout(t *testing.T) {

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  Counter
		args    args
		want    Checkout
		wantErr bool
	}{
		{
			name: "New order",
			fields: Counter{
				Account: reader.MockNewFtcAccount(enum.AccountKindFtc),
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
					Price:    faker.PriceStdYear.Original,
					Discount: faker.PriceStdYear.PromotionOffer,
				},
				Payable: price.Charge{
					Amount:   258,
					Currency: "cny",
				},
				LiveMode: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
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

func TestPaymentConfig_order(t *testing.T) {

	type args struct {
		checkout Checkout
	}
	tests := []struct {
		name    string
		fields  Counter
		args    args
		wantErr bool
	}{
		{
			name: "New order",
			fields: Counter{
				Account: reader.MockNewFtcAccount(enum.AccountKindFtc),
				Price:   faker.PriceStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				checkout: Checkout{
					Kind: enum.OrderKindCreate,
					Cart: cart.Cart{
						Price:    faker.PriceStdYear.Original,
						Discount: faker.PriceStdYear.PromotionOffer,
					},
					Payable: price.Charge{
						Amount:   258,
						Currency: "cny",
					},
					LiveMode: true,
				},
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
			got, err := c.order(tt.args.checkout)
			if (err != nil) != tt.wantErr {
				t.Errorf("order() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got.ID)
			assert.Equal(t, got.Price, 258.0)
			assert.Equal(t, got.Amount, 258.0)
			assert.Equal(t, got.Kind, enum.OrderKindCreate)
		})
	}
}

func TestPaymentConfig_BuildIntent(t *testing.T) {

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  Counter
		args    args
		want    PaymentIntent
		wantErr bool
	}{
		{
			name: "New order",
			fields: Counter{
				Account: reader.MockNewFtcAccount(enum.AccountKindFtc),
				Price:   faker.PriceStdYear,
				Method:  enum.PayMethodAli,
				WxAppID: null.String{},
			},
			args: args{
				m: reader.Membership{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			got, err := c.BuildIntent(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("order() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotZero(t, got.Checkout)
		})
	}
}
