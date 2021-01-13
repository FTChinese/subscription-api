package product

// Duration is the number of billing cycles and trial days a subscription purchased.
type Duration struct {
	CycleCount int64 `json:"cycleCount" db:"cycle_count"`
	ExtraDays  int64 `json:"extraDays" db:"extra_days"`
}

func (d Duration) Plus(other Duration) Duration {
	return Duration{
		CycleCount: d.CycleCount + other.CycleCount,
		ExtraDays:  d.ExtraDays + other.ExtraDays,
	}
}
