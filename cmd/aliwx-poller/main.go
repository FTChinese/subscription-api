package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/poll"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/robfig/cron/v3"
	"log"
	"os"
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

func main() {
	logger := config.MustGetLogger(production)
	rwdMyDB := db.MustNewMySQL(config.MustMySQLAPIConn(production))

	log.Println("Launching ali-wx poller...")

	poller := poll.NewOrderPoller(rwdMyDB, logger)
	defer poller.Close()

	if run {
		err := poller.Start(false)
		if err != nil {
			log.Println(err)
		}

		return
	}

	// Schedule chronicle task.
	c := cron.New(cron.WithLocation(chrono.TZShanghai))

	_, err := c.AddFunc("@daily", func() {
		err := poller.Start(false)
		if err != nil {
			logger.Error(err.Error())
			_ = logger.Sync()
		}
	})

	if err != nil {
		log.Println(err)
		return
	}

	c.Run()
}
