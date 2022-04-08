package invoice

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"reflect"
	"testing"
	"time"
)

func Test_consumeAddOn(t *testing.T) {
	inv1 := Invoice{}

	inv2 := Invoice{}

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
			got := consumeAddOn(tt.args.addOns, tt.args.start)
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func Test_reduceInvoices(t *testing.T) {

	type args struct {
		invs []Invoice
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Reduce invoices days",
			args: args{
				invs: []Invoice{},
			},
			want: 366*3 + 3,
		},
		{
			name: "Reduce nil",
			args: args{invs: nil},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reduceInvoices(tt.args.invs); got != tt.want {
				t.Errorf("reduceInvoices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOnGroup_ToAddOn(t *testing.T) {

	tests := []struct {
		name string
		g    AddOnGroup
		want addon.AddOn
	}{
		{
			name: "Sum addon days",
			g: AddOnGroup{
				enum.TierStandard: {},
				enum.TierPremium:  {},
			},
			want: addon.AddOn{
				Standard: 367 * 2,
				Premium:  367,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.ToAddOn(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToAddOn() = %v, want %v", got, tt.want)
			}
		})
	}
}
