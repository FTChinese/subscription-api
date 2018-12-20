package wxlogin

import (
	"html/template"
	"strings"

	"gitlab.com/ftchinese/subscription-api/postoffice"
)

// EmailBody is used to
type EmailBody struct {
	UserName   string
	Email      string
	NickName   string
	MutexMerge bool // True indicates that one of the membership is expired while the other is not.
	FTCMember  Membership
	WxMember   Membership
}

// ComposeParcel geneate a Parcel to be send to user after accounts bound.
func (e EmailBody) ComposeParcel() (postoffice.Parcel, error) {
	tmpl, err := template.New("notification").Parse(postoffice.LetterBound)

	if err != nil {
		logger.WithField("trace", "ComposeEmail").Error(err)

		return postoffice.Parcel{}, err
	}

	var body strings.Builder
	err = tmpl.Execute(&body, e)

	if err != nil {
		logger.WithField("trace", "ComposeEmail").Error(err)
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToAddress:   e.Email,
		ToName:      e.UserName,
		Subject:     "绑定账号",
		Body:        body.String(),
	}, nil
}
