//go:build !production
// +build !production

package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

func MockNewSubs() Subs {

	subs, err := NewSubs(faker.MustGenStripeSubs(), ids.UserIDs{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		panic(err)
	}

	return subs
}

type MockSubsBuilder struct {
	ftcID   string
	edition price.StripeEdition
	status  enum.SubsStatus
}

func NewMockSubsBuilder(ftcID string) MockSubsBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockSubsBuilder{
		ftcID:   ftcID,
		edition: price.StripeEditions.MustFindByEdition(price.StdYearEdition, false),
		status:  enum.SubsStatusActive,
	}
}

func (b MockSubsBuilder) WithEdition(e price.Edition) MockSubsBuilder {
	b.edition = price.StripeEditions.MustFindByEdition(e, false)
	return b
}

func (b MockSubsBuilder) WithStatus(s enum.SubsStatus) MockSubsBuilder {
	b.status = s
	return b
}

func (b MockSubsBuilder) WithCanceled() MockSubsBuilder {
	return b.WithStatus(enum.SubsStatusCanceled)
}

func (b MockSubsBuilder) Build() Subs {
	start := time.Now()
	end := dt.NewTimeRange(start).WithCycle(b.edition.Cycle).End
	canceled := time.Time{}

	if b.status == enum.SubsStatusCanceled {
		end = time.Now().AddDate(0, 0, -1)
		start = dt.NewTimeRange(end).WithCycleN(b.edition.Cycle, -1).End
		canceled = end
	}

	return Subs{
		ID:                   faker.GenStripeSubID(),
		Edition:              b.edition.Edition,
		WillCancelAtUtc:      chrono.Time{},
		CancelAtPeriodEnd:    false,
		CanceledUTC:          chrono.TimeFrom(canceled), // Set it for automatic cancel.
		CurrentPeriodEnd:     chrono.TimeFrom(end),
		CurrentPeriodStart:   chrono.TimeFrom(start),
		CustomerID:           faker.GenCustomerID(),
		DefaultPaymentMethod: null.String{},
		SubsItem: SubsItem{
			ItemID:  faker.GenStripeItemID(),
			PriceID: b.edition.PriceID,
		},
		LatestInvoiceID: faker.GenInvoiceID(),
		LiveMode:        false,
		StartDateUTC:    chrono.TimeNow(),
		EndedUTC:        chrono.Time{},
		CreatedUTC:      chrono.TimeNow(),
		UpdatedUTC:      chrono.TimeNow(),
		Status:          b.status,
		FtcUserID:       null.StringFrom(b.ftcID),
		PaymentIntent: PaymentIntent{
			ID:                 faker.GenPaymentIntentID(),
			Amount:             0,
			CanceledAtUTC:      chrono.Time{},
			CancellationReason: "",
			ClientSecret:       null.String{},
			CreatedUtc:         chrono.Time{},
			Currency:           "",
			CustomerID:         "",
			InvoiceID:          "",
			LiveMode:           false,
			PaymentMethodID:    "",
			Status:             "",
		},
	}
}
