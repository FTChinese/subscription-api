package subs

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"testing"
)

func TestNewConfirmationResult(t *testing.T) {
	ftcID := uuid.New().String()

	order := NewMockOrderBuilder(ftcID).Build()
	pr := MockNewPaymentResult(order)

	type args struct {
		p ConfirmationParams
	}
	tests := []struct {
		name    string
		args    args
		want    ConfirmationResult
		wantErr bool
	}{
		{
			name: "New confirmation result",
			args: args{
				p: ConfirmationParams{
					Payment: pr,
					Order:   order,
					Member:  reader.Membership{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfirmationResult(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfirmationResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("NewConfirmationResult() got = %v, want %v", got, tt.want)
			//}

			t.Logf("Payment result %s", faker.MustMarshalIndent(got.Payment))
			t.Logf("Order %s", faker.MustMarshalIndent(got.Order))
			t.Logf("Invoices %s", faker.MustMarshalIndent(got.Invoices))
			t.Logf("Membership %s", faker.MustMarshalIndent(got.Membership))
			t.Logf("Snapshot %s", faker.MustMarshalIndent(got.Snapshot))
		})
	}
}
