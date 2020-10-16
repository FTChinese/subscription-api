package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripeSdk "github.com/stripe/stripe-go"
)

type SubsResult struct {
	StripeSubs *stripeSdk.Subscription
	Member     reader.Membership
	Snapshot   reader.MemberSnapshot
}
