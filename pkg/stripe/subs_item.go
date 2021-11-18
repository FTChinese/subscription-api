package stripe

import "github.com/stripe/stripe-go/v72"

// SubsItem extracts a subscription item id and price id
// from the first element of items data array.
// Usually one subscription has only one item.
type SubsItem struct {
	ItemID  string    `json:"subsItemId" db:"subs_item_id"`
	PriceID string    `json:"priceId" db:"price_id"`
	Price   PriceJSON `json:"price" db:"price"`
}

// NewSubsItem gets the subscription item id and price id from a stripe subscription.
// stripe.Subscription.Items contains a list of subscription
// items, each with an attached price.
// See https://stripe.com/docs/api/subscriptions/object#subscription_object-items
// It the items Data array is empty, then it has nothing subscribed to.
func NewSubsItem(items *stripe.SubscriptionItemList) SubsItem {
	if items == nil || len(items.Data) == 0 {
		return SubsItem{}
	}

	return SubsItem{
		ItemID:  items.Data[0].ID,
		PriceID: items.Data[0].Price.ID,
		Price: PriceJSON{
			Price: NewPrice(items.Data[0].Price),
		},
	}
}
