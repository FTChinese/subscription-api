# Paywall

## Endpoints

* GET `/paywall?live=<bool>&refresh=<bool>` Loads the paywall data, either for live mode by specifying query parameter `live=true`, or sandbox mode with `live=false`. Default is true. Use `refresh=true` to bust cache.
* GET `/paywall/__migrate/active_prices` Use table subs_product.product_active_price to store which price is currently visible on paywall.
* POST `/paywall/banner` Create a banner
* POST `/paywall/banner/promo` Create a new promo
* DELETE `/paywall/banner/promo` Deelete latest promo

Products:

* GET `/paywall/products` List products
* POST `/payall/products` Create a new product.
* GET `/paywall/products/{id}` Load a product.
* PATCH `/paywall/products/{id}` Update a product
* POST `/paywall/products/{id}/activate` Activate a product.

Prices of a Product:

* GET `/paywall/prices?product_id=<string>` Get a list of prices belonging to a product
* POST `/paywall/prices` Create a price for product.
* POST `/paywall/prices/{id}/activate` Activate a price.
* POST `/paywall/prices/{id}/deactivate` Deactivate a price
* PATCH `/paywall/prices/{id}` Update a price
* PATCH `/paywall/prices/{id}/discounts` Refresh the offer list attached to a price.
* DELETE `/paywall/prices/{id}` Archive a price.

Offers of a price:

* GET `/paywall/discounts?price_id=<string>` List discounts under a price.
* POST `/paywall/discounts` Create a new discount for a price.
* GET `/paywall/discounts/{id}` Load a discount.
* DELETE `/paywall/discounts/{id}` Drop a discount.

## Database

Explains the usage of each table and how the paywall JSON data is constructed from them.

All tables have a `live_mode` column. If the value is `true`, if means this row is used for production; otherwise used for sandbox.

### subs_product.paywall_doc

Used to hold `banner` and `promo` fields in the JSON doc. These fields are saved as JSON in MySQL. Every modification will generate a new row based on last row.

### subs_product.paywall_product_v4

Stores product id currently visible on paywall. You can created as many products as like, but only two of them are visible on paywall. A compound unique key is used used to ensure no duplicate products are active simutaneously: `tier` and `live_mode` columns.

This table is deprecated. The `is_active` column in `product` table is used for the same purpose.

### subs_product.product

Each row is a product. You can create as many products as you like. For products of the same tier in a live/sandbox mode, only one of them is used on paywall.

### subs_product.price

A price must be attached to a product. A product could have multiple prices.

Cycle column might be null as introductory prices does not have it.

`is_active` field was used to tell which price is active on paywall under a product. However, a more robust approach is adopted: table `product_active_price` stores all prices visible on paywall.

### subs_product.product_active_price

A product can have multiple prices visible to end users. However, we must ensure the cycle field of thoese prices is unique. It is achieved by the `id` field in this table. The value of `id` is hexdecimal string of MD5 hash calculated from these fields of a price:

* a constant string `ftc` or `stripe`
* `tier` column
* `cycle` column
* `kind` column
* `live_mod` column

When retrieving paywall data, use a JOIN statement on this table and price table to get the unqiue prices under a product.

### subs_product.discount

Discount offers under a price. To avoid complicated JOIN statements, we put all active discounts of a price under the `discount_list` column in the `price` table. This is why there's an endpoint `/paywall/prices/{id}/discounts` to refresh a price's discount: it simply retrieves all active discount of a price and update the `discount_list` column.

### Tables deprecated

* subs_product.paywall_promo
* subs_product.paywall_banner

These two tables are replaced by a single table `paywall_doc`

* subs_product.paywall_product, replaced by `paywall_product_v4`
* subs_product.product_active_plans

* subs_product.plan, replaced by table `price`
