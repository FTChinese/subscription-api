package wxlogin

import (
	"html/template"
	"strings"

	"gitlab.com/ftchinese/subscription-api/postoffice"
)

// Send a notification email after user bound accounts.
const notificationEmail = `
FT中文网用户 {{.UserName}},

您好！您的FT中文网账号 {{.Email}} 已经绑定了微信账号 {{.NickName}}。

{{if .MutexMerge}}
您曾经用两个账号均购买过FT中文网会员服务，

{{if .FTCMember.IsExpired}}
其中用FT中文网账号购买的会员已于{{.FTCMember.ExpireDate}}过期，微信账号购买的会员尚未到期（到期日{{.WxMember.ExpireDate}}），
{{end}}

{{if .WxMember.IsExpired}}
其中用FT中文网账号购买的会员尚未到期（到期日{{.FTCMember.ExpireDate}}），微信账号购买的会员已于{{.WxMember.ExpireDate}}过期，
{{end}}

合并后您的会员期限和会员类型以未到期的订阅为准。
{{end}}

如有疑问，请联系客服：subscriber.service@ftchinese.com。

感谢您对FT中文网的支持。

FT中文网`

// EmailData is used to
type EmailData struct {
	UserName             string
	Email                string
	NickName             string
	MergedOneValidMember bool // True indicates that one of the membership is expired while the other is not. In such situtation tell user that memberships are merged; otherwise just tell user account merged without mentionning memberships.
	FTCMember            Membership
	WxMember             Membership
}

// ComposeParcel geneate a Parcel to be send to user after accounts bound.
func (e EmailData) ComposeParcel() (postoffice.Parcel, error) {
	tmpl, err := template.New("notification").Parse(notificationEmail)

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
