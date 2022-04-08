package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

const StmtShoppingSession = `
INSERT INTO premium.stripe_shopping_session
SET ftc_user_id = :ftc_user_id,
	recurring_price = :recurring_price,
	introductory_price = :introductory_price,
	membership = :membership,
	created_utc = :created_utc
`

// ShoppingSession is used to record the details when user
// creates/updates subscription.
type ShoppingSession struct {
	FtcUserID         string                  `db:"ftc_user_id"`
	RecurringPrice    PriceColumn             `db:"recurring_price"`
	IntroductoryPrice PriceColumn             `db:"introductory_price"`
	Membership        reader.MembershipColumn `db:"membership"`
	Intent            pw.CheckoutIntent       `db:"checkout_intent"`
	RequestParams     SubsReqParamsColumn     `db:"request_parameters"`
	CreatedUTC        chrono.Time             `db:"created_utc"`
}

func NewShoppingSession(cart pw.ShoppingCart, params pw.StripeSubsParams) ShoppingSession {
	return ShoppingSession{
		FtcUserID: cart.Account.FtcID,
		RecurringPrice: PriceColumn{
			StripePrice: cart.StripeItem.Recurring,
		},
		IntroductoryPrice: PriceColumn{
			StripePrice: cart.StripeItem.Introductory,
		},
		Membership: reader.MembershipColumn{
			Membership: cart.CurrentMember,
		},
		Intent:        cart.Intent,
		RequestParams: SubsReqParamsColumn{params},
		CreatedUTC:    chrono.TimeNow(),
	}
}

type SubsReqParamsColumn struct {
	pw.StripeSubsParams
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p SubsReqParamsColumn) Value() (driver.Value, error) {

	b, err := json.Marshal(p.StripeSubsParams)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (p *SubsReqParamsColumn) Scan(src interface{}) error {
	if src == nil {
		*p = SubsReqParamsColumn{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp pw.StripeSubsParams
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = SubsReqParamsColumn{tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to SubsReqParamsColumn")
	}
}
