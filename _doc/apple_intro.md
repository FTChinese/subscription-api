## Endpoints

No `X-User-Id` header is required unless explicitly specified.

The `{id}` placeholder here refers to Apple's original transaction id.

* POST `/apple/verify-receipt` Verify a receipt and returns the latest version.
* POST `/apple/link` Link email account to Apple subscription
* POST `/apple/unlink` Unlink ftc account from Apple subscription
* POST `/apple/subs` Verify a receipt and returns a condensed version of Apple's response.
* GET `/apple/subs` Get a list of a user's subscription. `X-User-Id` is required to identify the user.
* GET `/apple/subs/{id}` Load a single subscription
* PATCH `/apple/subs/{id}` Refresh an existing subscription.
* GET `/apple/recept/{id}` Load a receipt and its associated subscription.
