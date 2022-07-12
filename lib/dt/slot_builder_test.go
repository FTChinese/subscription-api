package dt

import (
	"reflect"
	"testing"
	"time"
)

func TestNewTimeRange(t *testing.T) {
	now := time.Now()

	type args struct {
		start time.Time
	}
	tests := []struct {
		name string
		args args
		want SlotBuilder
	}{
		{
			name: "New Date Range Instance",
			args: args{
				start: now,
			},
			want: SlotBuilder{
				Start: now,
				End:   now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSlotBuilder(tt.args.start); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
