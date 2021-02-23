package addon

import (
	"reflect"
	"testing"
)

func TestReservedDays_Plus(t *testing.T) {
	type fields struct {
		Standard int64
		Premium  int64
	}
	type args struct {
		other ReservedDays
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ReservedDays
	}{
		{
			name: "Plus",
			fields: fields{
				Standard: 15,
				Premium:  23,
			},
			args: args{
				other: ReservedDays{
					Standard: 366,
					Premium:  0,
				},
			},
			want: ReservedDays{
				Standard: 15 + 366,
				Premium:  23 + 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ReservedDays{
				Standard: tt.fields.Standard,
				Premium:  tt.fields.Premium,
			}
			if got := d.Plus(tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Plus() = %v, want %v", got, tt.want)
			}
		})
	}
}
