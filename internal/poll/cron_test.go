package poll

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/go-co-op/gocron"
	"testing"
)

func TestCron(t *testing.T) {
	s := gocron.NewScheduler(chrono.TZShanghai)

	_, err := s.Every(1).Seconds().Do(func() {
		t.Log("I am a running task")
	})

	if err != nil {
		t.Log(err)
		return
	}

	s.StartBlocking()
}
