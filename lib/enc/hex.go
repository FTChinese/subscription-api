package enc

import (
	"crypto/md5"
	"encoding/hex"
)

// HexStringSum returns a hexdecimal string of a string's md5 checksum.
// You are recommended to use VARBINARY(16) to save it in SQL.
// Two hex chars requires 8-bit (1 byte).
// 128-bits md5 checksum requires 128/8 = 16 bytes.
func HexStringSum(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
