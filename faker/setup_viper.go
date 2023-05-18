package faker

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/FTChinese/subscription-api/pkg/config"
)

func ReadConfigFile() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(home, "config", "env.dev.toml"))
}

func MustReadConfigFile() []byte {
	b, err := ReadConfigFile()
	if err != nil {
		panic(err)
	}

	return b
}

func MustSetupViper() {
	config.MustSetupViper(MustReadConfigFile())
}
