package access

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/conv"
	"github.com/guregu/null"
)

// OAuth contains the data related to an access token, used
// either by human or machines.
type OAuth struct {
	Token     conv.HexBin `gorm:"column:access_token"`
	Active    bool        `gorm:"column:is_active"`
	ExpiresIn null.Int    `gorm:"column:expires_in"` // seconds
	CreatedAt chrono.Time `gorm:"column:created_utc"`
}

func (o OAuth) TableName() string {
	return "oauth.access"
}

func (o OAuth) Expired() bool {

	if o.ExpiresIn.IsZero() {
		return false
	}

	expireAt := o.CreatedAt.Add(time.Second * time.Duration(o.ExpiresIn.Int64))

	return expireAt.Before(time.Now())
}
