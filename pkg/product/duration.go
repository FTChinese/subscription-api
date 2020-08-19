package product

// Duration is the number of billing cycles and trial period a subscription plan provides.
type Duration struct {
	CycleCount int64 `json:"cycleCount" db:"cycle_count"`
	ExtraDays  int64 `json:"extraDays" db:"extra_days"`
}

// DefaultDuration specifies the default value for a duration
func DefaultDuration() Duration {
	return Duration{
		CycleCount: 1,
		ExtraDays:  1,
	}
}
