package account

import (
	"github.com/FTChinese/go-rest/chrono"
	"time"
)

type SignUpRateParams struct {
	IP      string
	StartAt chrono.Time
	EndAt   chrono.Time
}

// NewSignUpRateParams creates the parameters used to check
// whether the specified ip has too many signup.
// By default we set the duration to 1 hour.
// For every hour the same ip could not signup
// for more than 60 times
func NewSignUpRateParams(ip string, h int64) SignUpRateParams {
	now := time.Now()
	endAt := chrono.TimeFrom(now)
	startAt := chrono.TimeFrom(now.Add(time.Duration(-h) * time.Hour))

	return SignUpRateParams{
		IP:      ip,
		StartAt: startAt,
		EndAt:   endAt,
	}
}

// SignUpLimit calculates how many new accounts are created at the same IP within the specified duration.
type SignUpLimit struct {
	Count int64 `db:"su_count"`
}

func (l SignUpLimit) Exceeds() bool {
	return l.Count > 60
}
