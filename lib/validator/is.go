package validator

import "regexp"

func IsMobile(m string) bool {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, m)

	return matched
}
