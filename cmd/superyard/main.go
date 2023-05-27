package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/FTChinese/subscription-api/pkg/config"
)

//go:embed build/api.toml
var tomlConfig string

//go:embed client_version_next
var clientVersionNext string

var clientVersionNg string

var (
	isProduction bool
	version      string
	build        string
)

func init() {
	flag.BoolVar(&isProduction, "production", false, "Indicate productions environment if present")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	config.MustSetupViper([]byte(tomlConfig))
}

func main() {

}
