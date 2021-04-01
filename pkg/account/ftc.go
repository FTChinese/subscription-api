package account

import "github.com/guregu/null"

// Ftc contains the minimal information to identify a user.
type Ftc struct {
	FtcID      string      `json:"id" db:"ftc_id"`           // FTC's uuid
	UnionID    null.String `json:"unionId" db:"wx_union_id"` // Wechat's union id
	StripeID   null.String `json:"stripeId" db:"stripe_id"`  // Stripe's id
	Email      string      `json:"email" db:"email"`         // Required, unique. Max 64.
	Password   string      `json:"-" db:"password"`
	Mobile     null.String `json:"mobile" db:"mobile_phone"`
	UserName   null.String `json:"userName" db:"user_name"` // Required, unique. Max 64.
	AvatarURL  null.String `json:"avatarUrl" db:"ftc_avatar_url"`
	IsVerified bool        `json:"isVerified" db:"is_verified"`
}
