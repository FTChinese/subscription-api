package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/subscription-api/internal"
	"github.com/FTChinese/subscription-api/pkg/config"
	"log"
	"os"
)

var (
	version    string
	build      string
	commit     string
	production bool
	sandbox    bool
)

func init() {
	flag.BoolVar(&production, "production", false, "Connect to production MySQL database if present. Default to localhost.")
	flag.BoolVar(&sandbox, "sandbox", false, "Use sandbox for alipay and wxpay webhook url and stripe keys")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	config.MustSetupViper()
}

func main() {

	s := internal.ServerStatus{
		Version:    version,
		Build:      build,
		Commit:     commit,
		Port:       "8202", // Version 1 uses port 8201
		Production: production,
		Sandbox:    sandbox,
	}

	log.Printf("Starting subscription api %s, built at %s with commit %s", s.Version, s.Build, s.Commit)

	log.Printf("Production %t. Sandbox %t. Port %s", s.Production, s.Sandbox, s.Port)

	internal.StartServer(s)
}
