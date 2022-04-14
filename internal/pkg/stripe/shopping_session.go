package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

const StmtShoppingSession = `
INSERT INTO premium.stripe_shopping_session
SET ftc_user_id = :ftc_user_id,
	recurring_price = :recurring_price,
	introductory_price = :introductory_price,
	checkout_intent = :checkout_intent,
	coupon = :coupon,
	membership = :membership,
	request_parameters = :request_parameters,
	subs_id = :subs_id,
	created_utc = :created_utc
`

// ShoppingSession is used to record the details when user
// creates/updates subscription.
type ShoppingSession struct {
	FtcUserID         string                  `db:"ftc_user_id"`
	RecurringPrice    PriceColumn             `db:"recurring_price"`
	IntroductoryPrice PriceColumn             `db:"introductory_price"`
	Coupon            CouponColumn            `db:"coupon"`
	Intent            reader.CheckoutIntent   `db:"checkout_intent"`
	Membership        reader.MembershipColumn `db:"membership"`
	RequestParams     SubsReqParamsColumn     `db:"request_parameters"`
	SubsID            null.String             `db:"subs_id"` // Only exists after subscription success
	CreatedUTC        chrono.Time             `db:"created_utc"`
}

func NewShoppingSession(cart reader.ShoppingCart, params SubsParams) ShoppingSession {
	return ShoppingSession{
		FtcUserID:         cart.Account.FtcID,
		RecurringPrice:    PriceColumn{cart.StripeItem.Recurring},
		IntroductoryPrice: PriceColumn{cart.StripeItem.Introductory},
		Coupon:            CouponColumn{cart.StripeItem.Coupon},
		Membership:        reader.MembershipColumn{Membership: cart.CurrentMember},
		Intent:            cart.Intent,
		RequestParams:     SubsReqParamsColumn{params},
		CreatedUTC:        chrono.TimeNow(),
	}
}

func (s ShoppingSession) WithSubs(id string) ShoppingSession {
	s.SubsID = null.StringFrom(id)

	return s
}

type SubsReqParamsColumn struct {
	SubsParams
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (p SubsReqParamsColumn) Value() (driver.Value, error) {

	b, err := json.Marshal(p.SubsParams)
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
		var tmp SubsParams
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
