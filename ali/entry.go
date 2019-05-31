package ali

// EntryKind enumerates the platforms from where user called Alipay.
type EntryKind int

const (
	EntryDesktopWeb EntryKind = iota
	EntryMobileWeb
	EntryApp
)

func (k EntryKind) String() string {
	names := [...]string{
		"desktop",
		"mobile",
		"app",
	}

	if k < EntryDesktopWeb || k > EntryApp {
		return ""
	}

	return names[k]
}
