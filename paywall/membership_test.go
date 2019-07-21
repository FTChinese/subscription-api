package paywall

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"testing"

	"github.com/stripe/stripe-go"
)

func TestMembership_FromStripe(t *testing.T) {
	s := stripe.Subscription{}

	if err := json.Unmarshal([]byte(subData), &s); err != nil {
		t.Error(err)
	}

	id := uuid.New().String()

	type fields struct {
		member Membership
	}
	type args struct {
		id  UserID
		sub StripeSub
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "New member",
			fields: fields{
				member: Membership{},
			},
			args: args{
				id: UserID{
					CompoundID: id,
					FtcID:      null.StringFrom(id),
				},
				sub: NewStripeSub(&s),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.member

			got, err := m.FromStripe(tt.args.id, tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.FromStripe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Stripe member: %+v", got)
		})
	}
}
