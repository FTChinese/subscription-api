package products

import (
	"github.com/FTChinese/subscription-api/pkg/pw"
)

// retrieveActiveProducts retrieve all products present on paywall.
func (env Env) retrieveActiveProducts() ([]pw.ProductBody, error) {
	var products = make([]pw.ProductBody, 0)

	err := env.db.Select(&products, pw.StmtPaywallProducts)

	if err != nil {
		return nil, err
	}

	return products, nil
}

type productsResult struct {
	value []pw.ProductBody
	error error
}

// asyncRetrieveProducts retrieves products in a goroutine.
func (env Env) asyncRetrieveProducts() <-chan productsResult {
	ch := make(chan productsResult)

	go func() {
		products, err := env.retrieveActiveProducts()

		ch <- productsResult{
			value: products,
			error: err,
		}
	}()

	return ch
}
