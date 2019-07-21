package paywall

import (
	"github.com/pkg/errors"
	"strings"
	"text/template"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/guregu/null"
)

// UserID is used to identify an FTC user.
// A user might have an ftc uuid, or a wechat union id,
// or both.
// This type structure is used to ensure unique constraint
// for SQL columns that cannot be both null since SQL do not
// have a mechanism to do UNIQUE INDEX on two columns while
// keeping either of them nullable.
// A user's compound id is taken from either ftc uuid or
// wechat id, with ftc id taking precedence.
type UserID struct {
	CompoundID string      `json:"-"`
	FtcID      null.String `json:"-"`
	UnionID    null.String `json:"-"`
}

func NewID(ftcID, unionID string) (UserID, error) {
	id := UserID{
		FtcID:   null.NewString(ftcID, ftcID != ""),
		UnionID: null.NewString(unionID, unionID != ""),
	}

	if ftcID != "" {
		id.CompoundID = ftcID
	} else if unionID != "" {
		id.CompoundID = unionID
	} else {
		return id, errors.New("ftcID and unionID should not both be null")
	}
	return id, nil
}

// FtcUser represents a row retrieve from userinfo table.
type FtcUser struct {
	UserID   string
	UnionID  null.String
	StripeID null.String
	Email    string
	UserName null.String
}

func (u FtcUser) ID() UserID {
	return UserID{
		CompoundID: u.UserID,
		FtcID:      null.StringFrom(u.UserID),
		UnionID:    u.UnionID,
	}
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (u FtcUser) NormalizeName() string {
	if u.UserName.Valid {
		return strings.Split(u.UserName.String, "@")[0]
	}

	return strings.Split(u.Email, "@")[0]
}

// ConfirmationParcel create a parcel for email after subscription is confirmed.
func (u FtcUser) ConfirmationParcel(s Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterNewSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User FtcUser
		Subs Subscription
	}{
		u,
		s,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.NormalizeName(),
		Subject:     "会员订阅",
		Body:        body.String(),
	}, nil
}

func (u FtcUser) StripeSubParcel(s StripeSub) (postoffice.Parcel, error) {
	tmpl, err := template.New("stripe_sub").Parse(letterStripeSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, _ := s.BuildFtcPlan()
	data := struct {
		User FtcUser
		Sub  StripeSub
		Plan Plan
	}{
		User: u,
		Sub:  s,
		Plan: plan,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.NormalizeName(),
		Subject:     "Stripe订阅",
		Body:        body.String(),
	}, nil
}

func (u FtcUser) StripeInvoiceParcel(i EmailedInvoice) (postoffice.Parcel, error) {
	tmpl, err := template.New("stripe_invoice").Parse(letterStripeInvoice)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, _ := i.BuildFtcPlan()
	data := struct {
		User    FtcUser
		Invoice EmailedInvoice
		Plan    Plan
	}{
		User:    u,
		Invoice: i,
		Plan:    plan,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.NormalizeName(),
		Subject:     "Stripe订阅发票",
		Body:        body.String(),
	}, nil
}
