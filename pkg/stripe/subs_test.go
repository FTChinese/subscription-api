package stripe

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"testing"
)

func TestNewSubs(t *testing.T) {
	var ss stripe.Subscription
	if err := json.Unmarshal([]byte(faker.StripeSubs), &ss); err != nil {
		t.Error(err)
		return
	}

	subs, err := NewSubs(&ss, pkg.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s", faker.MustMarshalIndent(subs))
}

func TestSubs_IsAutoRenewal(t *testing.T) {
	subs := MockNewSubs()

	tests := []struct {
		name   string
		fields Subs
		want   bool
	}{
		{
			name:   "Is auto renewal",
			fields: subs,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Subs{
				ID:                   tt.fields.ID,
				Edition:              tt.fields.Edition,
				WillCancelAtUtc:      tt.fields.WillCancelAtUtc,
				CancelAtPeriodEnd:    tt.fields.CancelAtPeriodEnd,
				CanceledUTC:          tt.fields.CanceledUTC,
				CurrentPeriodEnd:     tt.fields.CurrentPeriodEnd,
				CurrentPeriodStart:   tt.fields.CurrentPeriodStart,
				CustomerID:           tt.fields.CustomerID,
				DefaultPaymentMethod: tt.fields.DefaultPaymentMethod,
				SubsItem:             tt.fields.SubsItem,
				LatestInvoiceID:      tt.fields.LatestInvoiceID,
				LiveMode:             tt.fields.LiveMode,
				StartDateUTC:         tt.fields.StartDateUTC,
				EndedUTC:             tt.fields.EndedUTC,
				CreatedUTC:           tt.fields.CreatedUTC,
				UpdatedUTC:           tt.fields.UpdatedUTC,
				Status:               tt.fields.Status,
				FtcUserID:            tt.fields.FtcUserID,
			}
			if got := s.IsAutoRenewal(); got != tt.want {
				t.Errorf("IsAutoRenewal() = %v, want %v", got, tt.want)
			}
		})
	}
}
