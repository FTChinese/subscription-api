package addon

import (
	"github.com/FTChinese/go-rest/enum"
	"reflect"
	"testing"
)

func TestReservedDays_Plus(t *testing.T) {
	type fields struct {
		Standard int64
		Premium  int64
	}
	type args struct {
		other AddOn
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   AddOn
	}{
		{
			name: "Plus",
			fields: fields{
				Standard: 15,
				Premium:  23,
			},
			args: args{
				other: AddOn{
					Standard: 366,
					Premium:  0,
				},
			},
			want: AddOn{
				Standard: 15 + 366,
				Premium:  23 + 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := AddOn{
				Standard: tt.fields.Standard,
				Premium:  tt.fields.Premium,
			}
			if got := d.Plus(tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Plus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReservedDays_Clear(t *testing.T) {
	type fields struct {
		Standard int64
		Premium  int64
	}
	type args struct {
		tier enum.Tier
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   AddOn
	}{
		{
			name: "Clean Standard",
			fields: fields{
				Standard: 10,
				Premium:  5,
			},
			args: args{
				tier: enum.TierStandard,
			},
			want: AddOn{
				Standard: 0,
				Premium:  5,
			},
		},
		{
			name: "Clean Premium",
			fields: fields{
				Standard: 10,
				Premium:  5,
			},
			args: args{
				tier: enum.TierPremium,
			},
			want: AddOn{
				Standard: 10,
				Premium:  0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := AddOn{
				Standard: tt.fields.Standard,
				Premium:  tt.fields.Premium,
			}
			if got := d.Clear(tt.args.tier); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clear() = %v, want %v", got, tt.want)
			}
		})
	}
}
