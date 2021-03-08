package invoice

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

func TestNewAddOnGroup(t *testing.T) {
	userID := uuid.New().String()
	inv1 := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()
	inv2 := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithPrice(faker.PricePrm).
		Build()
	inv3 := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCarryOver).
		Build()

	type args struct {
		inv []Invoice
	}
	tests := []struct {
		name string
		args args
		want AddOnGroup
	}{
		{
			name: "Group Add-ons",
			args: args{
				inv: []Invoice{
					inv1,
					inv2,
					inv3,
				},
			},
			want: AddOnGroup{
				enum.TierStandard: []Invoice{
					inv1,
					inv3,
				},
				enum.TierPremium: []Invoice{
					inv2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAddOnGroup(tt.args.inv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddOnGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consumeAddOn(t *testing.T) {
	userID := uuid.New().String()
	inv1 := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()

	inv2 := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCarryOver).
		Build()

	now := time.Now()

	type args struct {
		addOns []Invoice
		start  time.Time
	}
	tests := []struct {
		name string
		args args
		want []Invoice
	}{
		{
			name: "Consume add-on",
			args: args{
				addOns: []Invoice{
					inv1,
					inv2,
				},
				start: now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConsumeAddOn(tt.args.addOns, tt.args.start)
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
