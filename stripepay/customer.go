package stripepay

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/card"
	"github.com/stripe/stripe-go/customer"
)

// CreateCustomer calls Stripe's Create a customer API.
// See https://stripe.com/docs/api/customers/create.
// NOTE: every time you hit this endpoint, Stripe creates
// a new customer, even with the same email. Avoid creating duplicate customer for the same user.
func CreateCustomer(email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)

	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

// GetCustomerCards list a customer's payment sources
// See https://stripe.com/docs/api/customers/retrieve
func GetCustomerCards(customerID string) (*stripe.SourceList, error) {
	c, err := customer.Get(customerID, nil)

	// Example error:
	// {
	//  "error": {
	//    "code": "resource_missing",
	//    "doc_url": "https://stripe.com/docs/error-codes/resource-missing",
	//    "message": "No such customer: cus_FM1UYLYALKBLth",
	//    "param": "id",
	//    "type": "invalid_request_error"
	//  }
	//}
	if err != nil {
		return nil, err
	}

	return c.Sources, nil
}

// AddCard adds a card to a customer.
func AddCard(customerID string, token string) (*stripe.Card, error) {
	params := &stripe.CardParams{
		Customer: stripe.String(customerID),
		Token:    stripe.String(token),
	}

	c, err := card.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}
