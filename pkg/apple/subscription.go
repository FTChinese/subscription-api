package apple

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

type SubsFilter struct {
	ExpiresIn   int64  // expires_date_utc
	AutoRenewal bool   // Allow null
	OrderBy     string // Order by column: updated_utc, expires_date_utc
	Ascending   bool
}

// Subscription contains a user's subscription data.
// It it built from app store's verification response.
// The original transaction id is used to uniquely identify a user.
type Subscription struct {
	BaseSchema
	LastTransactionID string      `json:"lastTransactionId" db:"last_transaction_id"`
	ProductID         string      `json:"productId" db:"product_id"`
	PurchaseDateUTC   chrono.Time `json:"purchaseDateUtc" db:"purchase_date_utc"`
	ExpiresDateUTC    chrono.Time `json:"expiresDateUtc" db:"expires_date_utc"`
	product.Edition
	AutoRenewal bool        `json:"autoRenewal" db:"auto_renewal"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC  chrono.Time `json:"updatedUtc" db:"updated_utc"`
}

// Membership build ftc's membership based on subscription
// from apple if this subscription is not linked
// to an ftc account.
func (s Subscription) NewMembership(id reader.MemberID) reader.Membership {
	m := reader.Membership{
		MemberID: id,
		Edition: product.Edition{
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
	}

	return m
}

// BuildOn updates an existing IAP membership based on this
// transaction.
func (s Subscription) BuildOn(m reader.Membership) reader.Membership {
	m.Tier = s.Tier
	m.Cycle = s.Cycle
	m.ExpireDate = chrono.DateFrom(s.ExpiresDateUTC.Time)
	m.PaymentMethod = enum.PayMethodApple
	m.FtcPlanID = null.String{}
	m.StripeSubsID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = s.AutoRenewal
	m.Status = enum.SubsStatusNull
	m.AppleSubsID = null.StringFrom(s.OriginalTransactionID)
	m.B2BLicenceID = null.String{}

	return m
}

type SubsList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Subscription `json:"data"`
	Err  error          `json:"-"`
}
