package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
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
		want TimeRange
	}{
		{
			name: "New Date Range Instance",
			args: args{
				start: now,
			},
			want: TimeRange{
				Start: now,
				End:   now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTimeRange(tt.args.start); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeRange_WithCycle(t *testing.T) {
	now := time.Now()

	type fields struct {
		Start time.Time
		End   time.Time
	}
	type args struct {
		cycle enum.Cycle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DatePeriod
	}{
		{
			name: "With yearly cycle",
			fields: fields{
				Start: now,
				End:   now,
			},
			args: args{
				cycle: enum.CycleYear,
			},
			want: DatePeriod{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(1, 0, 0)),
			},
		},
		{
			name: "With monthly cycle",
			fields: fields{
				Start: now,
				End:   now,
			},
			args: args{
				cycle: enum.CycleMonth,
			},
			want: DatePeriod{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(0, 1, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := TimeRange{
				Start: tt.fields.Start,
				End:   tt.fields.End,
			}
			if got := d.WithCycle(tt.args.cycle); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateRange_WithCycleN(t *testing.T) {
	now := time.Now()

	type fields struct {
		Start time.Time
		End   time.Time
	}
	type args struct {
		cycle enum.Cycle
		n     int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DatePeriod
	}{
		{
			name: "With 3 Years",
			fields: fields{
				Start: now,
				End:   now,
			},
			args: args{
				cycle: enum.CycleYear,
				n:     3,
			},
			want: DatePeriod{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(3, 0, 0)),
			},
		},
		{
			name: "With 3 Months",
			fields: fields{
				Start: now,
				End:   now,
			},
			args: args{
				cycle: enum.CycleMonth,
				n:     3,
			},
			want: DatePeriod{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(0, 3, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := TimeRange{
				Start: tt.fields.Start,
				End:   tt.fields.End,
			}
			if got := d.WithCycleN(tt.args.cycle, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCycleN() = %v, want %v", got, tt.want)
			}
		})
	}
}
