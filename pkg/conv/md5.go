package conv

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5SUM is a container for bytes of a string's md5 checksum.
// It could then converted to various representations.
// You are recommended to use VARBINARY(16) to save it in SQL.
// Two hex chars requires 8-bit (1 byte).
// 128-bits md5 checksum requires 128/8 = 16 bytes.
type MD5Sum []byte

func NewMD5Sum(s string) MD5Sum {
	h := md5.Sum([]byte(s))
	return h[:]
}

func (m MD5Sum) String() string {
	return hex.EncodeToString(m)
}

// ToHexBin change container to HexBin
// so that it could be used in JSON and SQL.
func (m MD5Sum) ToHexBin() HexBin {
	return HexBin(m)
}
