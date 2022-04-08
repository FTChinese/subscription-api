package pw

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
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
	CurrentMember reader.Membership
	Intent        CheckoutIntent
}

func NewShoppingCart(account account.BaseAccount) ShoppingCart {
	return ShoppingCart{
		Account:       account,
		FtcItem:       CartItemFtc{},
		StripeItem:    CartItemStripe{},
		CurrentMember: reader.Membership{},
		Intent:        CheckoutIntent{},
	}
}

func (s ShoppingCart) WithStripeItem(item CartItemStripe) ShoppingCart {
	s.FtcItem = CartItemFtc{}
	s.StripeItem = item

	return s
}

// WithFtcItem puts an FTC item into cart and clears stripe item.
func (s ShoppingCart) WithFtcItem(item CartItemFtc) ShoppingCart {
	s.FtcItem = item
	s.StripeItem = CartItemStripe{}

	return s
}

// WithMember sets current membership.
func (s ShoppingCart) WithMember(m reader.Membership) ShoppingCart {

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
			Kind:  reader.SubsKindForbidden,
			Error: errors.New("no purchase item set yet"),
		}
	}

	s.CurrentMember = m
	s.Intent = intent

	return s
}
