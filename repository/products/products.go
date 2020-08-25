package products

import "github.com/FTChinese/subscription-api/pkg/product"

// retrieveActiveProducts retrieve all products present on paywall.
func (env Env) retrieveActiveProducts() ([]product.Product, error) {
	var products = make([]product.Product, 0)

	err := env.db.Select(&products, product.StmtPaywallProducts)

	if err != nil {
		return nil, err
	}

	return products, nil
}

type productsResult struct {
	value []product.Product
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
