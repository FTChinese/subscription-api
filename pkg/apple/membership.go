package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

func NewMembership(userID pkg.UserIDs, s Subscription) reader.Membership {
	return reader.Membership{
		UserIDs: userID,
		Edition: price.Edition{
			Tier:  s.Tier,
			Cycle: s.Cycle,
		},
		ExpireDate:    chrono.DateFrom(s.ExpiresDateUTC.Time),
		PaymentMethod: enum.PayMethodApple,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   s.AutoRenewal,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.StringFrom(s.OriginalTransactionID),
		B2BLicenceID:  null.String{},
	}.Sync()
}
