package reader

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/guregu/null"
)

// ShoppingCart contains the essential information of a
// purchase attempt:
// - who want to purchase?
// - what item it is purchasing?
// - what is its current subscription status?
// FtcItem and StripeItem should be mutually exclusive.
type ShoppingCart struct {
	Account       account.BaseAccount
	FtcItem       CartItemFtc
	StripeItem    CartItemStripe
	PayMethod     enum.PayMethod
	WxAppID       null.String
	CurrentMember Membership
	Intent        CheckoutIntent
}

func NewShoppingCart(account account.BaseAccount) ShoppingCart {
	return ShoppingCart{
		Account:       account,
		FtcItem:       CartItemFtc{},
		StripeItem:    CartItemStripe{},
		CurrentMember: Membership{},
		Intent:        CheckoutIntent{},
	}
}

func (s ShoppingCart) WithStripeItem(item CartItemStripe) ShoppingCart {
	s.FtcItem = CartItemFtc{}
	s.StripeItem = item
	s.PayMethod = enum.PayMethodStripe

	return s
}

// WithFtcItem puts an FTC item into cart and clears stripe item.
func (s ShoppingCart) WithFtcItem(item CartItemFtc) ShoppingCart {
	s.FtcItem = item
	s.StripeItem = CartItemStripe{}

	return s
}

func (s ShoppingCart) WithAlipay() ShoppingCart {
	s.PayMethod = enum.PayMethodAli
	return s
}

func (s ShoppingCart) WithWxPay(appID string) ShoppingCart {
	s.PayMethod = enum.PayMethodWx
	s.WxAppID = null.StringFrom(appID)

	return s
}

// WithMember sets current membership.
func (s ShoppingCart) WithMember(m Membership) (ShoppingCart, error) {

	var intent CheckoutIntent

	if !s.FtcItem.Price.IsZero() {
		intent = NewCheckoutIntentFtc(
			m,
			s.FtcItem.Price)
	} else if !s.StripeItem.Recurring.IsZero() {
		intent = NewCheckoutIntentStripe(
			m,
			s.StripeItem.Recurring)
	} else {
		intent = CheckoutIntent{
			Kind:  IntentForbidden,
			Error: errors.New("no purchase item set yet"),
		}
	}

	s.CurrentMember = m
	s.Intent = intent

	return s, s.Intent.Error
}
