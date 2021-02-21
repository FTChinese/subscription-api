package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

type MembershipParams struct {
	UserIDs      reader.MemberID
	Subs         Subs
	ReservedDays addon.ReservedDays
}

func NewMembership(params MembershipParams) reader.Membership {

	expires := params.Subs.ExpiresAt()

	return reader.Membership{
		MemberID:      params.UserIDs,
		Edition:       params.Subs.Edition,
		LegacyTier:    null.IntFrom(reader.GetTierCode(params.Subs.Tier)),
		LegacyExpire:  null.IntFrom(expires.Unix()),
		ExpireDate:    chrono.DateFrom(expires),
		PaymentMethod: enum.PayMethodStripe,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.StringFrom(params.Subs.ID),
		StripePlanID:  null.StringFrom(params.Subs.PriceID),
		AutoRenewal:   params.Subs.IsAutoRenewal(),
		Status:        params.Subs.Status,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  params.ReservedDays,
	}
}
