package dt

import "time"

// PickLater select from two moments the one that comes later.
func PickLater(a time.Time, b time.Time) time.Time {
	if a.Equal(b) {
		return a
	}

	if a.Before(b) {
		return b
	}

	return a
}

// PickEarlier selects a time that comes earlier from two.
func PickEarlier(a time.Time, b time.Time) time.Time {
	if a.Equal(b) {
		return a
	}

	if a.Before(b) {
		return a
	}

	return b
}
