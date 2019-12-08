package subscription

// Duration is the number of billing cycles and extra days of a subscription plan provides.
type Duration struct {
	CycleCount int64 `json:"cycle_count" db:"cycle_count"`
	ExtraDays  int64 `json:"extra_days" db:"extra_days"`
}
