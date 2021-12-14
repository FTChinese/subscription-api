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

## Database change log

### v5.3.0

Tables in use

* subs_product.product
* subs_product.price. Replaces subs_product.plan. Cycle column might be null as introductory prices does not have it.
* subs_product.discount 
* subs_product.paywall_doc. Replaces paywall_promo and paywall_banner which became JSON columns of this table.
* ubs_product.paywall_product_v4. Replaces paywall_product. Added live_mode support.

Tables deprecated, kept only for backward compatibility:

* subs_product.paywall_product.
* subs_product.product_active_plans
* subs_product.paywall_promo
* subs_product.paywall_banner
* subs_product.plan

