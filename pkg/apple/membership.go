package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

type MembershipParams struct {
	UserID pkg.UserIDs // User id might comes from an existing membership when webhook message comes in, or from account is client is requesting link.
	Subs   Subscription
	AddOn  addon.AddOn
}

func NewMembership(params MembershipParams) reader.Membership {
	return reader.Membership{
		UserIDs:       params.UserID,
		Edition:       params.Subs.Edition,
		ExpireDate:    chrono.DateFrom(params.Subs.ExpiresDateUTC.Time),
		PaymentMethod: enum.PayMethodApple,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   params.Subs.AutoRenewal,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.StringFrom(params.Subs.OriginalTransactionID),
		B2BLicenceID:  null.String{},
		AddOn:         params.AddOn,
	}.Sync()
}
