package conv

import "strconv"

// FormatMoney converts human-used money amount to a string.
// Alipay uses this format.
func FormatMoney(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 32)
}

// MoneyCent converts human-used money amount to int64 in cent.
// Wechat pay uses this format.
func MoneyCent(n float64) int64 {
	return int64(n * 100)
}
