package account

import (
	"strings"

	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

const mobileEmailSuffix = "@ftchinese.user"

// MobileEmail generates a fake email from mobile.
func MobileEmail(m string) string {
	return m + mobileEmailSuffix
}

// BaseAccount contains the minimal information to identify a user.
// A user's account could have those possible combintations:
// * Email only
// * Mobile only
// * Wechat only
// * Email + Mobile
// * Email + Mobile + Wechat
// * Email + Wechat
type BaseAccount struct {
	FtcID        string      `json:"id" db:"ftc_id"`           // FTC's uuid
	UnionID      null.String `json:"unionId" db:"wx_union_id"` // Wechat's union id
	StripeID     null.String `json:"stripeId" db:"stripe_id"`  // Stripe's id
	Email        string      `json:"email" db:"email"`         // Required, unique. Max 64.
	Password     string      `json:"-" db:"password"`
	Mobile       null.String `json:"mobile" db:"mobile_phone"`
	UserName     null.String `json:"userName" db:"user_name"` // Required, unique. Max 64.
	AvatarURL    null.String `json:"avatarUrl" db:"ftc_avatar_url"`
	IsVerified   bool        `json:"isVerified" db:"is_verified"`
	CampaignCode null.String `json:"campaignCode" db:"campaign_code"`
}

// NewEmailBaseAccount creates an email-oriented account.
func NewEmailBaseAccount(params input.EmailSignUpParams) BaseAccount {
	return BaseAccount{
		FtcID:      uuid.New().String(),
		UnionID:    null.String{},
		StripeID:   null.String{},
		Email:      params.Email,
		Password:   params.Password,
		Mobile:     params.Mobile,
		UserName:   null.StringFrom(params.Email),
		AvatarURL:  null.String{},
		IsVerified: false,
	}
}

// NewMobileBaseAccount creates a mobile-oriented account.
func NewMobileBaseAccount(params input.MobileSignUpParams) BaseAccount {

	return BaseAccount{
		FtcID:      uuid.New().String(),
		UnionID:    null.String{},
		StripeID:   null.String{},
		Email:      MobileEmail(params.Mobile),
		Password:   rand.String(8),
		Mobile:     null.StringFrom(params.Mobile),
		UserName:   null.StringFrom(params.Mobile),
		AvatarURL:  null.String{},
		IsVerified: false,
	}
}

func (a BaseAccount) IsFtc() bool {
	return a.FtcID != ""
}

func (a BaseAccount) IsWxOnly() bool {
	return a.FtcID == "" && a.UnionID.Valid
}

// IsMobileEmail checks if a user's email is a faked one derived from phone number
// but Mobile field is missing.
func (a BaseAccount) IsMobileEmail() bool {
	return strings.HasSuffix(a.Email, mobileEmailSuffix) && a.Mobile.IsZero()
}

func (a BaseAccount) IsMobileOnly() bool {
	return strings.HasSuffix(a.Email, mobileEmailSuffix) && (a.Mobile.IsZero() || strings.HasPrefix(a.Email, a.Mobile.String))
}

func (a BaseAccount) IsStripeCustomer(targetID string) bool {
	if a.StripeID.IsZero() {
		return false
	}

	return a.StripeID.String == targetID
}

func (a BaseAccount) ExtractEmailName() string {
	return strings.Split(a.Email, "@")[0]
}

func (a BaseAccount) GetMobile() string {
	if a.Mobile.Valid {
		return a.Mobile.String
	}

	if strings.HasSuffix(a.Email, mobileEmailSuffix) {
		return a.ExtractEmailName()
	}

	return ""
}

func (a BaseAccount) SyncMobile() BaseAccount {
	m := a.ExtractEmailName()

	if !validator.IsMobile(m) {
		return a
	}

	a.Mobile = null.StringFrom(m)

	return a
}

func (a BaseAccount) WithUserName(name string) BaseAccount {
	a.UserName = null.StringFrom(name)

	return a
}

func (a BaseAccount) WithEmail(email string) BaseAccount {
	a.Email = email
	return a
}

func (a BaseAccount) WithMobile(m string) BaseAccount {
	a.Mobile = null.StringFrom(m)
	return a
}

func (a BaseAccount) WithStripeID(id string) BaseAccount {
	a.StripeID = null.StringFrom(id)

	return a
}

func (a BaseAccount) CompoundID() string {
	if a.FtcID != "" {
		return a.FtcID
	}

	if a.UnionID.Valid {
		return a.UnionID.String
	}

	return ""
}

func (a BaseAccount) CompoundIDs() ids.UserIDs {
	return ids.UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(a.FtcID, a.FtcID != ""),
		UnionID:    a.UnionID,
	}.MustNormalize()
}

func (a BaseAccount) IsTest() bool {
	return strings.HasSuffix(a.Email, ".test@ftchinese.com")
}

// NormalizeName returns username, or the name part of email if name does not exist.
func (a BaseAccount) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return a.ExtractEmailName()
}

// UnlinkWx is used when unlink an email-wechat account
func (a BaseAccount) UnlinkWx() BaseAccount {
	a.UnionID = null.String{}
	return a
}
