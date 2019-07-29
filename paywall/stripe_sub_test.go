package paywall

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/chrono"
	"testing"
	"time"

	"github.com/stripe/stripe-go"
)

func TestUnmarshalStripeSub(t *testing.T) {
	s := stripe.Subscription{}

	if err := json.Unmarshal([]byte(subDataExpanded), &s); err != nil {
		t.Error(err)
	}

	t.Logf("%d", s.EndedAt)
	t.Logf("%+v", s)

	t.Log(time.Unix(0, 0))
	t.Log(time.Unix(s.EndedAt, 0))

	t.Log(chrono.TimeFrom(time.Time{}))
}

func TestNewStripeSub(t *testing.T) {
	s := stripe.Subscription{}

	if err := json.Unmarshal([]byte(subDataExpanded), &s); err != nil {
		t.Error(err)
	}

	got := NewStripeSub(&s)

	t.Logf("%+v", got)
}

func TestNewStripeSubNotExpanded(t *testing.T) {
	s := stripe.Subscription{}

	if err := json.Unmarshal([]byte(subDataNotExpanded), &s); err != nil {
		t.Error(err)
	}

	t.Logf("Lates invoide: %+v", s.LatestInvoice)

	got := NewStripeSub(&s)

	t.Logf("%+v", got)
}
