package model

// Membership contains a user's membership details
type Membership struct {
	UserID string
	Tier   string
	Cycle  string
	Start  string
	Expire string
}
