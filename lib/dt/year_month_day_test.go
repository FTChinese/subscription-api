package dt

import (
	"github.com/FTChinese/go-rest/enum"
	"reflect"
	"testing"
)

func TestNewYearMonthDay(t *testing.T) {
	type args struct {
		cycle enum.Cycle
	}
	tests := []struct {
		name string
		args args
		want YearMonthDay
	}{
		{
			name: "New year",
			args: args{
				cycle: enum.CycleYear,
			},
			want: YearMonthDay{
				Years:  1,
				Months: 0,
				Days:   1,
			},
		},
		{
			name: "New month",
			args: args{
				cycle: enum.CycleMonth,
			},
			want: YearMonthDay{
				Years:  0,
				Months: 1,
				Days:   1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewYearMonthDay(tt.args.cycle); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewYearMonthDay() = %v, want %v", got, tt.want)
			}
		})
	}
}
