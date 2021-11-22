# Paywall

## Endpoints 

* GET `/paywall?live=<true|false>` Loads the paywall data, either for live mode by specifying query parameter `live=true`, or sandbox mode with `live=false`. Default is true.
* GET `/paywall/prices?product_id=<string>` Get a list of prices belonging to a product
* POST `/paywall/prices` Create a price for product.
* POST `/paywall/prices/{id}/activate` Activate a price.
* POST `/paywall/prices/{id}/refresh` Refresh a price so that its discounts updated.
* DELETE `/paywall/prices/{id}` Archive a price
* GET `/paywall/discounts?price_id=<string>` Get a list of discounts belong to a price.
* POST `/paywall/discounts` Create a new discount for a price.
* DELETE `/paywall/discounts/{id}` Drop a discount.
* GET `/paywall/__refresh` Bust cache.






