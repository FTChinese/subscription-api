package models

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"strconv"
	"strings"
)

var errAndroidForbidden = errors.New("您当前版本不支持订阅，请前往Play商店下载最新版。如果您无法访问Google Play，请勿购买。")

type SemVer struct {
	Major int
	Minor int
	Patch int
}

func ParseSemVer(v string) SemVer {
	var parts = make([]int, 0)

	for _, v := range strings.Split(v, ".") {
		n, err := strconv.Atoi(v)
		if err != nil {
			n = 0
		}

		parts = append(parts, n)
	}

	gap := len(parts) - 3
	if gap > 0 {
		for i := 0; i < gap; i++ {
			parts = append(parts, 0)
		}
	}

	return SemVer{
		Major: parts[0],
		Minor: parts[1],
		Patch: parts[2],
	}
}

// Compare compares two semantic versions.
// Returns negative number if s should come before (is smaller) other;
// 0 if the two are equal;
// Positive number if s should come after (is larger than) other.
func (s SemVer) Compare(other SemVer) int {
	diff := s.Major - other.Major
	if diff != 0 {
		return diff
	}

	diff = s.Minor - other.Minor
	if diff != 0 {
		return diff
	}

	return s.Patch - other.Patch
}

func (s SemVer) Equal(other SemVer) bool {
	return s.Compare(other) == 0
}

func (s SemVer) Larger(other SemVer) bool {
	return s.Compare(other) > 0
}

func (s SemVer) Smaller(other SemVer) bool {
	return s.Compare(other) < 0
}

var minimumVersion = SemVer{
	Major: 3,
	Minor: 1,
	Patch: 3,
}

// []string{"ftc", "play", "standard", "premium", "b2b"}
var permittedFlavors = map[string]struct{}{
	"ftc":      {},
	"play":     {},
	"standard": {},
	"premium":  {},
	"b2b":      {},
}

// AllowAndroidPurchase checks whether the Android client is
// permitted to buy membership.
func AllowAndroidPurchase(app util.ClientApp) error {

	if app.ClientType != enum.PlatformAndroid {
		return nil
	}

	if app.Version.IsZero() {
		return errAndroidForbidden
	}

	// Android version value is a string like `3.2.0-standard`
	// You have to extract each part.
	version := strings.Split(app.Version.String, "-")
	versionName := version[0]
	var flavor string
	if len(version) > 1 {
		flavor = version[1]
	}

	_, ok := permittedFlavors[flavor]
	if !ok {
		return errAndroidForbidden
	}

	semVer := ParseSemVer(versionName)

	if semVer.Smaller(minimumVersion) {
		return errAndroidForbidden
	}

	return nil
}
