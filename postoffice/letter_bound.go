package postoffice

// LetterBound is the content of the email send to user when FTC account is bound to a Wechat account.
// struct {
//     UserName
//     Email
//     NickName
//     MutexMerge
//     FTCMember Membership
//     WxMember Membership
// } {
//
// }
const LetterBound = `
尊敬的FT中文网用户 {{.UserName}},

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
