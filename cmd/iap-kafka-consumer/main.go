package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/subscription-api/internal/iap"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"os"
)

var (
	version    string
	build      string
	production bool
)

func init() {
	flag.BoolVar(&production, "production", false, "Connect to production MySQL database if present. Default to localhost.")
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

	myDB := db.MustNewMySQL(config.MustMySQLMasterConn(production))
	rdb := db.NewRedis(config.MustRedisAddress().Pick(production))

	c := iap.NewConsumer(myDB, rdb, logger, production)

	c.Consume()
}
