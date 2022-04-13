package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/stripe/stripe-go/v72"
)

// SubsItem extracts a subscription item id and price id
// from the first element of items data array.
// Usually one subscription has only one item.
type SubsItem struct {
	ID             string            `json:"id"`
	Price          price.StripePrice `json:"price"`
	Created        int64             `json:"created"`
	Quantity       int64             `json:"quantity"`
	SubscriptionID string            `json:"subscriptionId"`
}

// NewSubsItem gets the subscription item id and price id from a stripe subscription.
// stripe.Subscription.Items contain a list of subscription
// items, each with an attached price.
// See https://stripe.com/docs/api/subscriptions/object#subscription_object-items
// It the items Data array is empty, then it has nothing subscribed to.
func NewSubsItem(item *stripe.SubscriptionItem) SubsItem {
	return SubsItem{
		ID:             item.ID,
		Price:          price.NewStripePrice(item.Price),
		Created:        item.Created,
		Quantity:       item.Quantity,
		SubscriptionID: item.Subscription,
	}
}

// SubsItemList implements sql Value and Scan interface so that
// we could save a list of SubsItems as JSON.
type SubsItemList []SubsItem

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (l SubsItemList) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (l *SubsItemList) Scan(src interface{}) error {
	if src == nil {
		*l = nil
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp SubsItemList
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to PriceColumn")
	}
}

func NewSubsItemList(items *stripe.SubscriptionItemList) SubsItemList {
	var ret = make([]SubsItem, 0)

	for _, item := range items.Data {
		ret = append(ret, NewSubsItem(item))
	}

	return ret
}
