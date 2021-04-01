package account

import "github.com/guregu/null"

// Wechat contain the essential data to identify a wechat user.
type Wechat struct {
	WxNickname  null.String `json:"nickname" db:"wx_nickname"`
	WxAvatarURL null.String `json:"avatarUrl" db:"wx_avatar_url"`
}
