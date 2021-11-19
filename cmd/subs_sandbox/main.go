package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/FTChinese/subscription-api/internal"
	"github.com/FTChinese/subscription-api/pkg/config"
	"log"
	"os"
)

//go:embed build/api.toml
var tomlConfig string

var (
	version    string
	build      string
	commit     string
	production bool
	sandbox    bool
	liveMode   bool
)

func init() {
	flag.BoolVar(&production, "production", true, "Connect to production MySQL database if present. Default to localhost.")
	flag.BoolVar(&liveMode, "livemode", true, "Determine live/sandbox mode for webhook, and which of Stripe or Apple service to use")

	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	config.MustSetupViper([]byte(tomlConfig))
}

func main() {

	s := internal.ServerStatus{
		Version:    version,
		Build:      build,
		Commit:     commit,
		Port:       "8201", // Sandbox port never changes
		Production: production,
		LiveMode:   liveMode,
	}

	log.Printf("Starting subscription api %s, built at %s with commit %s", s.Version, s.Build, s.Commit)

	log.Printf("Production %t. LiveMode %t. Port %s", s.Production, s.LiveMode, s.Port)

	internal.StartServer(s)
}
