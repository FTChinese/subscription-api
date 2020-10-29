package wechat

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/spf13/viper"
	"testing"
)

func TestWxApps(t *testing.T) {
	config.MustSetupViper()
	var apps WxApps
	err := viper.UnmarshalKey("wxapp", &apps)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%v", apps)
}
