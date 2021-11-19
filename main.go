package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/FTChinese/subscription-api/internal"
	"github.com/FTChinese/subscription-api/pkg/config"
)

//go:embed build/api.toml
var tomlConfig string

var (
	version string
	build   string
	commit  string
	// For development, production == true, sandbox == true
	// For online sandbox server, production == true, sandbox == true
	// For online production server, production == true, sandbox == false
	production bool // Determine which db to use
	sandbox    bool // Deprecated. Determine the mode when using Stripe/Apple service.
	liveMode   bool // Determine the mode when using Stripe/Apple service.
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

// Version 1: Port 8200, https://www.ftacademy.cn/api/v1
// Version 2. Port 8202, subs-api-v2, https://www.ftacademy.cn/api/v2
// Version 3. Port 8203, subs-api-v3, https://www.ftacademy.cn/api/v3
func main() {

	s := internal.ServerStatus{
		Version:    version,
		Build:      build,
		Commit:     commit,
		Port:       config.Port,
		Production: production,
		LiveMode:   liveMode,
	}

	log.Printf("Starting subscription api %s, built at %s with commit %s", s.Version, s.Build, s.Commit)

	log.Printf("Production %t. LiveMode %t. Port %s", s.Production, s.LiveMode, s.Port)

	internal.StartServer(s)
}
