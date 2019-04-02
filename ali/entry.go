package ali

// EntryKind enumerates the platforms from where user called Alipay.
type EntryKind int

const (
	EntryDesktopWeb EntryKind = iota
	EntryMobileWeb
	EntryApp
)