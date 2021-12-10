package price

// Kind represents the type of price.
type Kind string

const (
	KindRecurring Kind = "recurring"
	KindOneTime   Kind = "one_time"
)
