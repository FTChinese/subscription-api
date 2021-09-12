package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/poll"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/go-co-op/gocron"
	"log"
	"os"
	"time"
)

var (
	version    string
	build      string
	production bool // Command line argument. Determine which db to use: true use production mysql, false use localhost.
	run        bool
)

func init() {
	flag.BoolVar(&production, "production", false, "Connect to production MySQL database if present. Default to localhost.")
	flag.BoolVar(&run, "run", false, "Run immediately")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	config.MustSetupViper()
}

func task() {
	log.Printf("Starting iap polling job at %s", time.Now().Format(time.RFC3339))

	logger := config.MustGetLogger(production)
	myDB := db.MustNewMyDBs(production)

	poller := poll.NewIAPPoller(myDB, production, logger)

	err := poller.Start(false)
	if err != nil {
		log.Println(err)
	}

	poller.Close()
}

func main() {
	log.Println("Launching IAP poller...")

	if run {
		task()
		return
	}

	s := gocron.NewScheduler(chrono.TZShanghai)
	_, err := s.Every(1).
		Day().
		At("00:00").
		Do(task)

	if err != nil {
		panic(err)
	}

	s.StartBlocking()
}
