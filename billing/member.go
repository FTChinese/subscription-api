package billing

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func genMmID() (string, error) {
	s, err := gorest.RandomBase64(9)
	if err != nil {
		return "", err
	}

	return "mmb_" + s, nil
}

// Membership contains a user's membership details
// This is actually called subscription by Stripe.
type Member struct {
	ID            null.String    `json:"id"`
	CompoundID    string         `json:"-"` // Either FTCUserID or UnionID
	FTCUserID     null.String    `json:"-"`
	UnionID       null.String    `json:"-"` // For both vip_id_alias and wx_union_id columns.
	Plan          Plan           `json:"plan"`
	ExpireDate    chrono.Date    `json:"expireDate"`
	PaymentMethod enum.PayMethod `json:"payMethod"`
	StripeSubID   null.String    `json:"-"`
	AutoRenewal   bool           `json:"autoRenewal"`
}
