package subs

// Duration is the number of billing cycles and extra days of a subscription plan provides.
type Duration struct {
	CycleCount int64 `json:"cycleCount" db:"cycle_count"`
	ExtraDays  int64 `json:"extraDays" db:"extra_days"`
}
