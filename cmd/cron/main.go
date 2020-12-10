package main

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

func main() {
	s := gocron.NewScheduler(chrono.TZShanghai)

	_, err := s.Every(1).Seconds().Do(func() {
		log.Printf("I am a running task at %d", time.Now().Unix())
	})

	if err != nil {
		panic(err)
	}

	s.StartBlocking()
}
