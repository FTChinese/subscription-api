package conv

import "strings"

func SlashConcat(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return strings.Join(strings.Split(s, " "), "-")
}
