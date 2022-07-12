package dt

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"time"
)

type SlotBuilder struct {
	Start time.Time
	End   time.Time
}

func NewSlotBuilder(start time.Time) SlotBuilder {
	return SlotBuilder{
		Start: start,
		End:   start,
	}
}

func (builder SlotBuilder) WithPeriod(d YearMonthDay) SlotBuilder {
	builder.End = builder.End.AddDate(int(d.Years), int(d.Months), int(d.Days))

	return builder
}

func (builder SlotBuilder) WithCycle(cycle enum.Cycle) SlotBuilder {
	return builder.WithPeriod(NewYearMonthDay(cycle))
}

// WithCycleN adds n cycles to end date.
func (builder SlotBuilder) WithCycleN(cycle enum.Cycle, n int) SlotBuilder {
	return builder.WithPeriod(NewYearMonthDayN(cycle, n))
}

func (builder SlotBuilder) AddYears(years int) SlotBuilder {
	builder.End = builder.End.AddDate(years, 0, 0)
	return builder
}

func (builder SlotBuilder) AddMonths(months int) SlotBuilder {
	builder.End = builder.End.AddDate(0, months, 0)
	return builder
}

func (builder SlotBuilder) AddDays(days int) SlotBuilder {
	builder.End = builder.End.AddDate(0, 0, days)

	return builder
}

func (builder SlotBuilder) StartTime() chrono.Time {
	return chrono.TimeFrom(builder.Start)
}

func (builder SlotBuilder) EndTime() chrono.Time {
	return chrono.TimeFrom(builder.End)
}

func (builder SlotBuilder) StartDate() chrono.Date {
	return chrono.DateFrom(builder.Start)
}

func (builder SlotBuilder) EndDate() chrono.Date {
	return chrono.DateFrom(builder.End)
}

func (builder SlotBuilder) Build() TimeSlot {
	return TimeSlot{
		StartUTC: chrono.TimeFrom(builder.Start),
		EndUTC:   chrono.TimeFrom(builder.End),
	}
}
