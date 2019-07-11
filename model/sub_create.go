package model

import "gitlab.com/ftchinese/subscription-api/paywall"

// NewSub creates a new subscription or renew a subscription.
// Those are grouped into new subscription:
// * New member
// * An expired member
//
// This does not allow upgrading to premium member if the
// member is expired.
// Upgrading: a member is not expired, and the product
// it is subscribed to is not premium.
func (env Env) NewSub() (paywall.Subscription, error) {

}
