package poll

import (
	"github.com/robfig/cron/v3"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	c := cron.New()

	_, err := c.AddFunc("* * * * *", func() {
		t.Logf("Hello %s", time.Now().Format(time.RFC3339))
	})

	if err != nil {
		t.Error(err)
		return
	}

	c.Run()
}
