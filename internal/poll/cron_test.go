package poll

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/go-co-op/gocron"
	"testing"
)

func task() {
	fmt.Print("I'm running")
}

func TestCron(t *testing.T) {
	s := gocron.NewScheduler(chrono.TZShanghai)

	_, err := s.Every(1).Second().Do(task)

	if err != nil {
		t.Log(err)
		return
	}

	s.StartBlocking()
}
