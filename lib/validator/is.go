package validator

import "regexp"

func IsMobile(m string) bool {
	matched, _ := regexp.MatchString(`^(1[3|4|5|8][0-9]\d{4,8})$`, m)

	return matched
}
