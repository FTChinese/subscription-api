package subs

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"testing"
)

func TestConfirmationParams_purchasedTimeParams(t *testing.T) {
	order := NewMockOrderBuilder("").Build()
	pr := MockNewPaymentResult(order)

	type fields struct {
		Payment PaymentResult
		Order   Order
		Member  reader.Membership
	}
	tests := []struct {
		name   string
		fields fields
		want   PurchasedTimeParams
	}{
		{
			name: "Parameter to deduce purchased time",
			fields: fields{
				Payment: pr,
				Order:   order,
				Member:  reader.Membership{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ConfirmationParams{
				Payment: tt.fields.Payment,
				Order:   tt.fields.Order,
				Member:  tt.fields.Member,
			}

			got := p.purchasedTimeParams()

			//if got := p.purchasedTimeParams(); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("purchasedTimeParams() = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
