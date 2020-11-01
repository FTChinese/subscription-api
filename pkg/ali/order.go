package ali

type OrderReq struct {
	Title       string
	FtcOrderID  string
	TotalAmount string
	WebhookURL  string
	TxKind      EntryKind
	ReturnURL   string // The callback url to redirect after paid in browser.
}
