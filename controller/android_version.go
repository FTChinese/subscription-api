package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"strings"
)

var errAndroidForbidden = errors.New("您当前版本不支持订阅，请前往Play商店下载最新版。如果您无法访问Google Play，请勿购买。")

const (
	vThreeOneThree = iota + 22
	vThreeOneFour
)

var androidVersions = map[string]int{
	"3.1.3": vThreeOneThree,
	"3.1.4": vThreeOneFour,
}

func getAndroidVersion(name string) (int, error) {
	versionCode, ok := androidVersions[name]
	if !ok {
		return 0, errAndroidForbidden
	}

	return versionCode, nil
}

func allowAndroidPurchase(app util.ClientApp) error {

	if app.ClientType != enum.PlatformAndroid {
		return nil
	}

	if app.Version.IsZero() {
		return errAndroidForbidden
	}

	version := strings.Split(app.Version.String, "-")
	versionName := version[0]
	var flavor string
	if len(version) > 1 {
		flavor = version[1]
	}

	if flavor != "ftc" && flavor != "play" {
		return errAndroidForbidden
	}

	versionCode, err := getAndroidVersion(versionName)
	if err != nil {
		return err
	}

	if versionCode < vThreeOneThree {
		return errAndroidForbidden
	}

	return nil
}
