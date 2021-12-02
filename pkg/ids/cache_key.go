package ids

var boolKey = map[bool]string{
	true:  "live",
	false: "test",
}

func GetBoolKey(k bool) string {
	return boolKey[k]
}

func PaywallCacheKey(live bool) string {
	return "paywall_" + GetBoolKey(live)
}
