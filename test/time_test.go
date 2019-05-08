package test

import (
	"testing"
	"time"
)

func TestUnixDate(t *testing.T) {
	now := time.Now()

	date := now.Truncate(24 * time.Hour)

	t.Logf("%s\n", date.In(time.UTC).Format(time.RFC3339))
	t.Logf("%d\n", date.Unix())
}
