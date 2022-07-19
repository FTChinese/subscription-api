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

func (cart ShoppingCart) WithStripeItem(item CartItemStripe) ShoppingCart {
	cart.FtcItem = CartItemFtc{}
	cart.StripeItem = item
	cart.PayMethod = enum.PayMethodStripe

	return cart
}

// WithFtcItem puts an FTC item into cart and clears stripe item.
func (cart ShoppingCart) WithFtcItem(item CartItemFtc) ShoppingCart {
	cart.FtcItem = item
	cart.StripeItem = CartItemStripe{}

	return cart
}

func (cart ShoppingCart) WithAlipay() ShoppingCart {
	cart.PayMethod = enum.PayMethodAli
	return cart
}

func (cart ShoppingCart) WithWxPay(appID string) ShoppingCart {
	cart.PayMethod = enum.PayMethodWx
	cart.WxAppID = null.StringFrom(appID)

	return cart
}

// WithMember sets current membership.
func (cart ShoppingCart) WithMember(m Membership) (ShoppingCart, error) {

	var intent CheckoutIntent

	// Build intent for ftc pay.
	if !cart.FtcItem.Price.IsZero() {
		intent = NewCheckoutIntentFtc(
			m,
			cart.FtcItem.Price)
	} else if !cart.StripeItem.Recurring.IsZero() {
		// Build intent for stripe pay.
		intent = NewCheckoutIntentStripe(m, cart.StripeItem)
	} else {
		intent = CheckoutIntent{
			Kind:  IntentForbidden,
			Error: errors.New("no purchase item set yet"),
		}
	}

	cart.CurrentMember = m
	cart.Intent = intent

	return cart, cart.Intent.Error
}
