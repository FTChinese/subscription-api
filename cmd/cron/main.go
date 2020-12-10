package main

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

func main() {
	s := gocron.NewScheduler(chrono.TZShanghai)

	log.Println("Cron job started")

	_, err := s.Every(1).Hour().Do(func() {
		log.Printf("I am a running task at %s", time.Now().Format(time.RFC3339))
	})

	if err != nil {
		panic(err)
	}

	s.StartBlocking()
}
