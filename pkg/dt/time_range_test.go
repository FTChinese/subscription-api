package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"reflect"
	"testing"
	"time"
)

func TestNewDateRange(t *testing.T) {
	now := time.Now()

	type args struct {
		start time.Time
	}
	tests := []struct {
		name string
		args args
		want DateRange
	}{
		{
			name: "New Date Range Instance",
			args: args{
				start: now,
			},
			want: DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDateRange(tt.args.start); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateRange_WithCycle(t *testing.T) {
	now := time.Now()

	type fields struct {
		StartDate chrono.Date
		EndDate   chrono.Date
	}
	type args struct {
		cycle enum.Cycle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DateRange
	}{
		{
			name: "With yearly cycle",
			fields: fields{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now),
			},
			args: args{
				cycle: enum.CycleYear,
			},
			want: DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(1, 0, 0)),
			},
		},
		{
			name: "With monthly cycle",
			fields: fields{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now),
			},
			args: args{
				cycle: enum.CycleMonth,
			},
			want: DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(0, 1, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DateRange{
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
			}
			if got := d.WithCycle(tt.args.cycle); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}
