package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

type OrderBuilder struct {
	id         reader.MemberID
	plan       plan.Plan
	method     enum.PayMethod
	k          SubsKind
	balance    float64
	wxAppID    string
	upIntentID string
	snapshotID string
}

func NewOrderBuilder(id reader.MemberID) *OrderBuilder {
	return &OrderBuilder{
		id: id,
	}
}

func (b *OrderBuilder) SetWxAppID(appID string) *OrderBuilder {
	b.wxAppID = appID
	return b
}

func (b *OrderBuilder) GetReaderID() reader.MemberID {
	return b.id
}

func (b *OrderBuilder) SetPlan(p plan.Plan) *OrderBuilder {
	b.plan = p
	return b
}

func (b *OrderBuilder) SetPayMethod(m enum.PayMethod) *OrderBuilder {
	b.method = m
	return b
}

// ----------
// The following parameters need to query db.

func (b *OrderBuilder) SetSubKind(k SubsKind) *OrderBuilder {
	b.k = k

	return b
}

// SetUpgradeIntentID set the foreign key point to an upgrade
// intent record.
func (b *OrderBuilder) SetUpgradeIntentID(id string) *OrderBuilder {
	b.upIntentID = id

	return b
}

// SetBalance if this is an upgrade order.
func (b *OrderBuilder) SetBalance(balance float64) *OrderBuilder {
	b.balance = balance

	return b
}

func (b *OrderBuilder) Build() (Order, error) {
	orderID, err := GenerateOrderID()
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:               orderID,
		MemberID:         b.id,
		Plan:             b.plan,
		Usage:            b.k,
		PaymentMethod:    b.method,
		WxAppID:          null.NewString(b.wxAppID, b.wxAppID != ""),
		StartDate:        chrono.Date{},
		EndDate:          chrono.Date{},
		CreatedAt:        chrono.TimeNow(),
		ConfirmedAt:      chrono.Time{},
		UpgradeIntentID:  null.NewString(b.upIntentID, b.upIntentID == ""),
		MemberSnapshotID: null.String{}, // Set when confirming order.
	}, nil
}
