package db

type WriteKind int

const (
	WriteKindDenial WriteKind = iota
	WriteKindInsert
	WriteKindUpdate
)
