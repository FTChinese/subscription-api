package test

// ID is use to determine how to generate user id.
type AccountKind int

const (
	AccountKindFtc AccountKind = iota
	AccountKindWx
	AccountKindLinked
)
