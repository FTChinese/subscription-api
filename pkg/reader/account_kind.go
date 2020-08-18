package reader

// AccountKind tells what kind of account reader is using.
type AccountKind int

const (
	AccountKindFtc    AccountKind = iota // FTC-only account
	AccountKindWx                        // Wx-only account
	AccountKindLinked                    // Linked account
)
